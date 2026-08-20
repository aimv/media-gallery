package cms

import "context"

type ContentBlockRepository interface {
	Save(ctx context.Context, block *ContentBlock) error
	GetByIDWithMedia(ctx context.Context, id string) (*BlockWithMedia, error)
	List(ctx context.Context) ([]*ContentBlock, error)
	Delete(ctx context.Context, id string) error
}
