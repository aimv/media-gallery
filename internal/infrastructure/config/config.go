// internal/infrastructure/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort        string
	DBDSN           string
	StorageRoot     string
	WorkerCount     int
	LeaseDuration   time.Duration
	MaxUploadSizeMB int64
	APIKey          string
	FFmpegPath      string
	FFprobePath     string
	LogLevel        string
}

func Load() (*Config, error) {
	// Загружаем .env, если файл существует.
	// В Docker Compose переменные окружения прокидываются напрямую.
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		DBDSN:           os.Getenv("DB_DSN"),
		StorageRoot:     os.Getenv("STORAGE_ROOT"),
		WorkerCount:     getEnvInt("WORKER_COUNT", 1),
		LeaseDuration:   getEnvDuration("LEASE_DURATION", 5*time.Minute),
		MaxUploadSizeMB: getEnvInt64("MAX_UPLOAD_SIZE_MB", 500),
		APIKey:          os.Getenv("API_KEY"),
		FFmpegPath:      getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:     getEnv("FFPROBE_PATH", "ffprobe"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DBDSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}
	if c.StorageRoot == "" {
		return fmt.Errorf("STORAGE_ROOT is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("API_KEY is required")
	}
	if c.WorkerCount < 1 {
		return fmt.Errorf("WORKER_COUNT must be >= 1")
	}
	if c.MaxUploadSizeMB <= 0 {
		return fmt.Errorf("MAX_UPLOAD_SIZE_MB must be > 0")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
