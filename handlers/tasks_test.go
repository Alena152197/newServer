package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"database/sql"
	"server_new/config"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) {
	// Используем тестовую БД в памяти
	var err error
	config.DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Ошибка открытия БД:", err)
	}

	// Создаём таблицы
	_, err = config.DB.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'pending',
			userid INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal("Ошибка создания таблицы:", err)
	}
}

func TestTasksHandler_CreateTask(t *testing.T) {
	setupTestDB(t)
	defer config.CloseDB()

	handler := NewTasksHandler()

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "валидная задача",
			body: map[string]interface{}{
				"title":       "Тест",
				"description": "Описание",
				"status":      "pending",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "пустой заголовок",
			body: map[string]interface{}{
				"title": "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-ID", "1")

			w := httptest.NewRecorder()
			handler.CreateTask(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("CreateTask() status = %v, want %v", w.Code, tt.expectedStatus)
			}
		})
	}
}
