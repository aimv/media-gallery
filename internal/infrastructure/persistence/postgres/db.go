package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт пул соединений с PostgreSQL, проверяет доступность БД
// с таймаутом 5 секунд и возвращает готовый к использованию пул.
func NewPool(dbDSN string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		return nil, fmt.Errorf("parse db dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	slog.Info("Connected to PostgreSQL", "max_conns", poolCfg.MaxConns)
	return pool, nil
}
