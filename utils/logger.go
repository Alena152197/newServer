package utils

import (
	"log/slog"
	"os"

	"server_new/config"
)

var Logger *slog.Logger

// InitLogger инициализирует логгер
func InitLogger() {
	// Читаем уровень из конфигурации
	logLevel := config.GetLogLevel()

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
}

// LogError логирует ошибку
func LogError(err error, msg string, args ...interface{}) {
	Logger.Error(msg, append([]interface{}{"error", err}, args...)...)
}

// LogInfo логирует информационное сообщение
func LogInfo(msg string, args ...interface{}) {
	Logger.Info(msg, args...)
}

// LogWarn логирует предупреждение
func LogWarn(msg string, args ...interface{}) {
	Logger.Warn(msg, args...)
}
