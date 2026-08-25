// Package config отвечает за загрузку конфигурации приложения
// из переменных окружения с поддержкой .env файла.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config содержит все параметры, необходимые для работы сервиса.
type Config struct {
	DBDSN          string // строка подключения к PostgreSQL
	HTTPServerPort string // порт HTTP API
	APIKey         string // ключ для авторизации запросов
	StorageDir     string // базовая директория для файлового хранилища
	MaxUploadSize  int64  // максимальный размер загружаемого файла в байтах
	LogLevel       string // уровень логирования: debug, info, warn, error
}

// Load загружает конфигурацию из .env файла (если он есть)
// и переменных окружения. При отсутствии .env файла ошибка игнорируется,
// значения берутся напрямую из среды.
func Load() (*Config, error) {
	// Пытаемся загрузить .env; если файла нет — не считаем это ошибкой.
	_ = godotenv.Load()

	// Читаем атомарные параметры СУБД для динамической сборки DSN
	dbUser := getEnv("DB_USER", "media_user")
	dbPass := getEnv("DB_PASSWORD", "media_pass")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "media_gallery")

	// Динамически склеиваем валидную строку подключения по стандарту URL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	cfg := &Config{
		DBDSN:          dsn,
		HTTPServerPort: getEnv("HTTP_SERVER_PORT", "8080"),
		APIKey:         getEnv("API_KEY", ""),
		StorageDir:     getEnv("STORAGE_DIR", "./storage"),
		MaxUploadSize:  getEnvInt64("MAX_UPLOAD_SIZE", 524288000), // 500 МБ
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if cfg.DBDSN == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или defaultValue, если переменная пуста.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// getEnvInt64 возвращает целочисленное значение переменной окружения
// или defaultValue, если переменная пуста или содержит нечисловое значение.
func getEnvInt64(key string, defaultValue int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultValue
}
