// Package logger предоставляет утилиты для инициализации структурированного логирования.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// InitLogger создаёт и настраивает структурированный JSON-логгер на основе slog.
// level — строка уровня логирования: "debug", "info", "warn" или "error" (регистр не важен).
// Возвращает настроенный *slog.Logger и устанавливает его как дефолтный через slog.SetDefault.
func InitLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
