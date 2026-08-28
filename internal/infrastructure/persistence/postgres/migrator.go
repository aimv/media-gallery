// Package postgres содержит реализации портов для работы с PostgreSQL.
package postgres

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Драйвер для интеграции мигратора с СУБД PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Драйвер для чтения миграций из локальной файловой системы
)

// RunMigrations применяет SQL-миграции из указанного источника (например, "file://...") и DSN.
func RunMigrations(dbDSN string, migrationsPath string) error {
	slog.Info("Running database migrations", "source", migrationsPath)

	m, err := migrate.New(migrationsPath, dbDSN)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	slog.Info("Database migrations applied successfully")
	return nil
}
