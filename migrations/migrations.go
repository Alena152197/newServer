package migrations

import (
	"database/sql"
	"embed"
	"io/fs"
)

//go:embed *.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	// Читаем все SQL файлы
	files, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return err
	}

	// Применяем каждую миграцию
	for _, file := range files {
		sql, err := migrationsFS.ReadFile(file.Name())
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(sql)); err != nil {
			return err
		}
	}

	return nil
}
