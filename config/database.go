package config

import (
	"database/sql"  // стандартная библиотека для работы с БД
	"fmt"           // для форматирования строк
	"log"           // для логов
	"os"            // для работы с файловой системой
	"path/filepath" // для работы с путями
	"time"          // для работы со временем

	"server_new/migrations" // для миграций БД

	_ "github.com/mattn/go-sqlite3" // драйвер SQLite (импортируем для регистрации)
)

var DB *sql.DB // глобальная переменная для соединения с БД

// InitDB инициализирует базу данных
func InitDB() error {
	// Используем путь из конфигурации (из .env или значения по умолчанию)
	dbPath := DBPath
	if dbPath == "" {
		dbPath = ".tmp/base.sqlite"
	}

	// Создаём директорию для базы данных, если её нет
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("не удалось создать папку для БД (%s): %v", dbDir, err)
		}
	}

	// Открываем соединение с базой данных
	// Если файла нет, SQLite создаст его автоматически
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть базу данных: %v", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return fmt.Errorf("не удалось подключиться к базе: %v", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Включаем поддержку внешних ключей (важно для связей между таблицами)
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("не удалось включить внешние ключи: %v", err)
	}

	// Запускаем миграции
	if err := migrations.RunMigrations(db); err != nil {
		return fmt.Errorf("не удалось применить миграции: %v", err)
	}

	DB = db
	log.Println("База данных инициализирована успешно")
	return nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
