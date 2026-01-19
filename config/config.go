package config

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	Port           string
	JWTSecret      string
	DBPath         string
	AllowedOrigins []string
)

// Load загружает переменные окружения
func Load() error {
	// Загружаем .env файл (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения системы")
	}

	Port = getEnv("PORT", "4000")
	JWTSecret = getEnv("JWT_SECRET", "dev-secret-change-me")
	DBPath = getEnv("DB_PATH", ".tmp/base.sqlite")

	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	AllowedOrigins = strings.Split(originsStr, ",")

	// Убираем пробелы
	for i := range AllowedOrigins {
		AllowedOrigins[i] = strings.TrimSpace(AllowedOrigins[i])
	}

	return nil
}

// getEnv получает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetLogLevel возвращает уровень логирования из конфигурации
func GetLogLevel() slog.Level {
	levelStr := getEnv("LOG_LEVEL", "info")
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
