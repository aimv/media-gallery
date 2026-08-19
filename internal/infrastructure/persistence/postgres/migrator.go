package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations применяет SQL-миграции из указанной директории.
func RunMigrations(dbDSN string, migrationsDir string) error {
	m, err := migrate.New(
		"file://"+migrationsDir,
		dbDSN,
	)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
