// Package postgres содержит реализации портов для работы с PostgreSQL.
package postgres

import (
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Драйвер для интеграции мигратора с СУБД PostgreSQL
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Драйвер для чтения миграций из локальной файловой системы
)

// RunMigrations применяет SQL-миграции из указанной директории.
func RunMigrations(dbDSN string) error {
	slog.Info("Running database migrations", "source", "file://internal/infrastructure/persistence/postgres/migrations")

	// TODO: В рамках MVP путь зафиксирован в инфраструктурном адаптере.
	// В качестве следующего шага техдолга запланирован перевод миграций на go:embed для автономности бинарника.

	m, err := migrate.New("file://internal/infrastructure/persistence/postgres/migrations", dbDSN)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	slog.Info("Database migrations applied successfully")
	return nil
}
