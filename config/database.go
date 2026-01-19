package config

import (
	"database/sql"  // стандартная библиотека для работы с БД
	"fmt"           // для форматирования строк
	"log"           // для логов
	"os"            // для работы с файловой системой
	"path/filepath" // для работы с путями
	"time"          // для работы со временем

	_ "github.com/mattn/go-sqlite3" // драйвер SQLite (импортируем для регистрации)
)

var DB *sql.DB // глобальная переменная для соединения с БД

// InitDB инициализирует базу данных
func InitDB() error {
	// Создаём папку .tmp, если её нет
	tmpDir := ".tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать папку .tmp: %v", err)
	}

	// Путь к файлу базы данных
	dbPath := filepath.Join(tmpDir, "base.sqlite")

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

	// Создаём таблицы
	if err := createTables(db); err != nil {
		return fmt.Errorf("не удалось создать таблицы: %v", err)
	}

	DB = db
	log.Println("База данных инициализирована успешно")
	return nil
}

// createTables создаёт таблицы users и tasks
func createTables(db *sql.DB) error {
	// Создаём таблицу users
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу users: %v", err)
	}
	log.Println("Таблица users создана")

	// Создаём таблицу tasks
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending' CHECK (
			status IN ('pending', 'in_progress', 'completed')
		),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		due_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		userid INTEGER NOT NULL DEFAULT 1,
		FOREIGN KEY(userid) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(createTasksTable)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу tasks: %v", err)
	}
	log.Println("Таблица tasks создана")

	return nil
}

// CloseDB закрывает соединение с базой данных
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
