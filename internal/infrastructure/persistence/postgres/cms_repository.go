// internal/infrastructure/persistence/postgres/cms_repository.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aimv/media-gallery/internal/domain/cms"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CMSRepository struct {
	pool *pgxpool.Pool
}

func NewCMSRepository(pool *pgxpool.Pool) *CMSRepository {
	return &CMSRepository{pool: pool}
}

// Save атомарно сохраняет контентный блок и его связи с медиафайлами.
func (r *CMSRepository) Save(ctx context.Context, block *cms.ContentBlock) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	dataJSON, err := json.Marshal(block.Data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO content_blocks (
			id, page_key, block_type, data, sort_order, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, now(), now())
		ON CONFLICT (id) DO UPDATE SET
			page_key = EXCLUDED.page_key,
			block_type = EXCLUDED.block_type,
			data = EXCLUDED.data,
			sort_order = EXCLUDED.sort_order,
			updated_at = now()
	`,
		block.ID,
		block.PageKey,
		block.BlockType,
		dataJSON,
		block.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("upsert content block: %w", err)
	}

	// Удаляем старые связи и вставляем новые
	if _, err := tx.Exec(ctx, `DELETE FROM content_block_media_links WHERE block_id = $1`, block.ID); err != nil {
		return fmt.Errorf("delete old media links: %w", err)
	}

	for i, mediaID := range block.MediaIDs {
		_, err := tx.Exec(ctx, `
			INSERT INTO content_block_media_links (block_id, asset_id, position)
			VALUES ($1, $2, $3)
		`, block.ID, mediaID, i+1)
		if err != nil {
			return fmt.Errorf("insert media link: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// GetByIDWithMedia возвращает блок вместе со всеми привязанными медиафайлами.
// Использует LEFT JOIN и агрегацию для формирования списка MediaIDs и MediaInfo.
func (r *CMSRepository) GetByIDWithMedia(ctx context.Context, id string) (*cms.BlockWithMedia, error) {
	var (
		result      cms.BlockWithMedia
		dataRaw     []byte
		mediaIDs    []string
		mediaNames  []string
		mediaTypes  []string
		mediaStatus []string
	)

	err := r.pool.QueryRow(ctx, `
		SELECT
			cb.id,
			cb.page_key,
			cb.block_type,
			cb.data,
			cb.sort_order,
			cb.created_at,
			cb.updated_at,
			COALESCE(array_agg(ma.id ORDER BY cml.position) FILTER (WHERE ma.id IS NOT NULL), '{}'),
			COALESCE(array_agg(ma.original_file_name ORDER BY cml.position) FILTER (WHERE ma.id IS NOT NULL), '{}'),
			COALESCE(array_agg(ma.media_type ORDER BY cml.position) FILTER (WHERE ma.id IS NOT NULL), '{}'),
			COALESCE(array_agg(ma.status ORDER BY cml.position) FILTER (WHERE ma.id IS NOT NULL), '{}')
		FROM content_blocks cb
		LEFT JOIN content_block_media_links cml ON cb.id = cml.block_id
		LEFT JOIN media_assets ma ON cml.asset_id = ma.id
		WHERE cb.id = $1
		GROUP BY cb.id
	`, id).Scan(
		&result.ID,
		&result.PageKey,
		&result.BlockType,
		&dataRaw,
		&result.SortOrder,
		&result.CreatedAt,
		&result.UpdatedAt,
		&mediaIDs,
		&mediaNames,
		&mediaTypes,
		&mediaStatus,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("content block not found: %w", err)
		}
		return nil, fmt.Errorf("select content block with media: %w", err)
	}

	if err := json.Unmarshal(dataRaw, &result.Data); err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}

	result.MediaIDs = mediaIDs
	result.Media = make([]cms.MediaInfo, 0, len(mediaIDs))
	for i := range mediaIDs {
		result.Media = append(result.Media, cms.MediaInfo{
			ID:               mediaIDs[i],
			OriginalFileName: mediaNames[i],
			MediaType:        mediaTypes[i],
			Status:           mediaStatus[i],
		})
	}

	return &result, nil
}

// List возвращает все контентные блоки без связанных медиафайлов.
func (r *CMSRepository) List(ctx context.Context) ([]*cms.ContentBlock, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, page_key, block_type, data, sort_order, created_at, updated_at
		FROM content_blocks
		ORDER BY page_key, sort_order
	`)
	if err != nil {
		return nil, fmt.Errorf("list content blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*cms.ContentBlock
	for rows.Next() {
		var (
			b       cms.ContentBlock
			dataRaw []byte
		)
		if err := rows.Scan(
			&b.ID,
			&b.PageKey,
			&b.BlockType,
			&dataRaw,
			&b.SortOrder,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content block: %w", err)
		}
		if err := json.Unmarshal(dataRaw, &b.Data); err != nil {
			return nil, fmt.Errorf("unmarshal data: %w", err)
		}
		// Медиа-связи не подгружаем; при необходимости можно расширить
		blocks = append(blocks, &b)
	}
	return blocks, rows.Err()
}

// Delete удаляет контентный блок и его связи.
func (r *CMSRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM content_block_media_links WHERE block_id = $1`, id); err != nil {
		return fmt.Errorf("delete media links: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM content_blocks WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete content block: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
