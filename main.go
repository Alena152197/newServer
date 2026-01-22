package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"server_new/config"     // наш пакет с конфигурацией БД
	"server_new/handlers"   // обработчики запросов
	"server_new/middleware" // middleware для аутентификации
	"server_new/models"     // модели данных
	"server_new/utils"      // утилиты для работы с токенами и паролями
)

type InfoResponse struct {
	Message string `json:"message"`
}

// Структура для задачи (соответствует таблице в БД)
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserID      int    `json:"userid"` // новое поле для ID пользователя
}

// Структура для ошибки
type ErrorResponse struct {
	Error string `json:"error"`
}

// Обработчик для маршрута GET /info
func infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Сервер работает на порту %s"}`, config.Port)
}

// Обработчик для маршрута POST /auth/register
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.SendError(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		middleware.SendValidationError(w, "Неверный формат JSON")
		return
	}

	// Валидация username
	if ok, msg := utils.ValidateUsername(requestData.Username); !ok {
		middleware.SendValidationError(w, msg)
		return
	}

	// Валидация email
	if !utils.ValidateEmail(requestData.Email) {
		middleware.SendValidationError(w, "Укажи корректный email")
		return
	}

	// Валидация password
	if ok, msg := utils.ValidatePassword(requestData.Password); !ok {
		middleware.SendValidationError(w, msg)
		return
	}

	// Нормализуем данные
	normalizedEmail := strings.TrimSpace(strings.ToLower(requestData.Email))
	cleanUsername := strings.TrimSpace(requestData.Username)
	rawPassword := strings.TrimSpace(requestData.Password)

	// Создаём хэш пароля
	hashedPassword, err := utils.HashPassword(rawPassword)
	if err != nil {
		log.Printf("Ошибка хэширования пароля: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	// Вставляем пользователя в базу данных
	result, err := config.DB.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		cleanUsername, normalizedEmail, hashedPassword,
	)
	if err != nil {
		// Проверяем, не дублируется ли email или username
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			sendError(w, http.StatusBadRequest, "Пользователь с таким именем или email уже существует")
			return
		}
		log.Printf("Ошибка вставки в БД: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось создать пользователя")
		return
	}

	// Получаем ID созданного пользователя
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Ошибка получения ID: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось получить ID пользователя")
		return
	}

	// Отправляем ответ без пароля
	response := models.UserResponse{
		ID:        int(id),
		Username:  cleanUsername,
		Email:     normalizedEmail,
		CreatedAt: time.Now(),
	}

	sendJSON(w, http.StatusCreated, response)
}

// Обработчик для маршрута POST /auth/login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var requestData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Читаем и парсим JSON
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	// Нормализуем email
	normalizedEmail := strings.TrimSpace(strings.ToLower(requestData.Email))
	rawPassword := strings.TrimSpace(requestData.Password)

	if normalizedEmail == "" || rawPassword == "" {
		sendError(w, http.StatusBadRequest, "Укажи email и пароль")
		return
	}

	// Ищем пользователя по email
	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, username, email, password, created_at FROM users WHERE email = ?",
		normalizedEmail,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		// Пользователь не найден или ошибка БД
		sendError(w, http.StatusBadRequest, "Неверные email или пароль")
		return
	}

	// Проверяем пароль
	if !utils.CheckPassword(rawPassword, user.Password) {
		sendError(w, http.StatusBadRequest, "Неверные email или пароль")
		return
	}

	// Создаём JWT токен
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		log.Printf("Ошибка создания токена: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось создать токен")
		return
	}

	// Отправляем ответ с токеном и данными пользователя
	response := map[string]interface{}{
		"jwt": token,
		"user": models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}

	sendJSON(w, http.StatusOK, response)
}

// Обработчик для маршрута POST /auth/reset-simple
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var requestData struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	normalizedEmail := strings.TrimSpace(strings.ToLower(requestData.Email))
	newPassword := strings.TrimSpace(requestData.NewPassword)

	if !strings.Contains(normalizedEmail, "@") {
		sendError(w, http.StatusBadRequest, "Укажи корректный email")
		return
	}

	if len(newPassword) < 6 {
		sendError(w, http.StatusBadRequest, "Новый пароль должен быть не короче 6 символов")
		return
	}

	// Проверяем, существует ли пользователь
	var userID int
	err = config.DB.QueryRow(
		"SELECT id FROM users WHERE email = ?",
		normalizedEmail,
	).Scan(&userID)

	if err != nil {
		sendError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}

	// Хэшируем новый пароль
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		log.Printf("Ошибка хэширования пароля: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось обновить пароль")
		return
	}

	// Обновляем пароль
	_, err = config.DB.Exec(
		"UPDATE users SET password = ? WHERE id = ?",
		hashedPassword, userID,
	)
	if err != nil {
		log.Printf("Ошибка обновления пароля: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось обновить пароль")
		return
	}

	sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Обработчик для маршрута GET /me (защищённый)
func meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	// Извлекаем userID из заголовка (установлен middleware)
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID == 0 {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Получаем данные пользователя из базы
	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)

	if err != nil {
		sendError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}

	// Отправляем ответ без пароля
	response := models.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	sendJSON(w, http.StatusOK, response)
}

// Обработчик для маршрута PUT /me (обновление профиля)
func mePutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Получаем текущего пользователя
	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, email, password FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		sendError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}

	var requestData struct {
		Email           string `json:"email"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	updates := []string{}
	params := []interface{}{}

	// Смена email
	if requestData.Email != "" {
		newEmail := strings.TrimSpace(strings.ToLower(requestData.Email))
		if newEmail != user.Email {
			if !strings.Contains(newEmail, "@") {
				sendError(w, http.StatusBadRequest, "Укажи корректный email")
				return
			}
			updates = append(updates, "email = ?")
			params = append(params, newEmail)
		}
	}

	// Смена пароля
	if requestData.NewPassword != "" {
		if requestData.CurrentPassword == "" {
			sendError(w, http.StatusBadRequest, "Укажи текущий пароль")
			return
		}

		// Проверяем текущий пароль
		if !utils.CheckPassword(requestData.CurrentPassword, user.Password) {
			sendError(w, http.StatusUnauthorized, "Текущий пароль неверен")
			return
		}

		if len(requestData.NewPassword) < 6 {
			sendError(w, http.StatusBadRequest, "Новый пароль должен быть не короче 6 символов")
			return
		}

		newHash, err := utils.HashPassword(requestData.NewPassword)
		if err != nil {
			log.Printf("Ошибка хэширования пароля: %v", err)
			sendError(w, http.StatusInternalServerError, "Не удалось обновить пароль")
			return
		}

		updates = append(updates, "password = ?")
		params = append(params, newHash)
	}

	if len(updates) == 0 {
		sendError(w, http.StatusBadRequest, "Нет данных для обновления")
		return
	}

	params = append(params, userID)

	sql := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err = config.DB.Exec(sql, params...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			sendError(w, http.StatusConflict, "Этот email уже занят")
		} else {
			log.Printf("Ошибка обновления профиля: %v", err)
			sendError(w, http.StatusInternalServerError, "Не удалось обновить профиль")
		}
		return
	}

	sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Обработчик для маршрута DELETE /me (удаление аккаунта)
func meDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Удаляем задачи пользователя (благодаря CASCADE это произойдёт автоматически,
	// но лучше сделать явно для ясности)
	_, err = config.DB.Exec("DELETE FROM tasks WHERE userid = ?", userID)
	if err != nil {
		log.Printf("Ошибка удаления задач: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось удалить задачи пользователя")
		return
	}

	// Удаляем пользователя
	result, err := config.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		log.Printf("Ошибка удаления пользователя: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось удалить аккаунт")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendError(w, http.StatusNotFound, "Пользователь не найден")
		return
	}

	// Отправляем статус 204 без тела
	w.WriteHeader(http.StatusNoContent)
}

// Универсальный обработчик для маршрута /me (GET, PUT и DELETE)
func meHandlerWrapper(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		meHandler(w, r)
	case http.MethodPut:
		mePutHandler(w, r)
	case http.MethodDelete:
		meDeleteHandler(w, r)
	default:
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
	}
}

// Обработчик для маршрута GET /tasks (только свои задачи)
func tasksGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	// Получаем userID из запроса (установлен middleware)
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Ключевая часть — фильтрация по владельцу
	rows, err := config.DB.Query(
		"SELECT id, title, description, status, userid, created_at FROM tasks WHERE userid = ? ORDER BY id DESC",
		userID,
	)
	if err != nil {
		log.Printf("Ошибка запроса к БД: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось получить список задач")
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var createdAt string
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.UserID, &createdAt)
		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []Task{}
	}

	sendJSON(w, http.StatusOK, tasks)
}

// Обработчик для маршрута POST /tasks (создать задачу текущему пользователю)
func tasksPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	// Получаем userID из запроса
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Проверяем заголовок Content-Type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		sendError(w, http.StatusBadRequest, "Ожидается Content-Type: application/json")
		return
	}

	// Структура для входящих данных
	var requestData struct {
		Title       string `json:"title"`
		Description string `json:"description"` // новое поле
		Status      string `json:"status"`      // новое поле
	}

	// Читаем тело запроса и парсим JSON
	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		middleware.SendValidationError(w, "Неверный формат JSON")
		return
	}

	// Валидация title
	if ok, msg := utils.ValidateTaskTitle(requestData.Title); !ok {
		middleware.SendValidationError(w, msg)
		return
	}

	// Устанавливаем значения по умолчанию
	description := strings.TrimSpace(requestData.Description)
	status := strings.TrimSpace(requestData.Status)
	if status == "" {
		status = "pending" // по умолчанию
	}

	// Валидация status (если передан)
	if requestData.Status != "" && !utils.ValidateTaskStatus(requestData.Status) {
		middleware.SendValidationError(w, "Статус должен быть: pending, in_progress или completed")
		return
	}

	title := strings.TrimSpace(requestData.Title)

	// Сохраняем задачу с userid текущего пользователя
	result, err := config.DB.Exec(
		"INSERT INTO tasks (title, description, status, userid) VALUES (?, ?, ?, ?)",
		title, description, status, userID,
	)
	if err != nil {
		log.Printf("Ошибка вставки в БД: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось создать задачу")
		return
	}

	// Получаем ID созданной задачи
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Ошибка получения ID: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось получить ID задачи")
		return
	}

	// Создаём ответ
	newTask := Task{
		ID:          int(id),
		Title:       title,
		Description: description,
		Status:      status,
		UserID:      userID,
	}

	// Отправляем ответ со статусом 201 (Created)
	sendJSON(w, http.StatusCreated, newTask)
}

// Обработчик для маршрута GET /tasks/:id (только своя задача)
func taskGetByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Извлекаем ID из пути
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		sendError(w, http.StatusBadRequest, "Неверный формат пути")
		return
	}

	taskID, err := strconv.Atoi(parts[2])
	if err != nil {
		sendError(w, http.StatusBadRequest, "ID должен быть числом")
		return
	}

	// Ищем задачу по ID И userid (важно: проверяем владельца!)
	var task Task
	var createdAt string
	err = config.DB.QueryRow(
		"SELECT id, title, description, status, userid, created_at FROM tasks WHERE id = ? AND userid = ?",
		taskID, userID,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.UserID, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			sendError(w, http.StatusNotFound, "Задача не найдена")
		} else {
			log.Printf("Ошибка запроса к БД: %v", err)
			sendError(w, http.StatusInternalServerError, "Не удалось получить задачу")
		}
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// Обработчик для маршрута PUT /tasks/:id (обновить только свою задачу)
func taskPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Извлекаем ID из пути
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		sendError(w, http.StatusBadRequest, "Неверный формат пути")
		return
	}

	taskID, err := strconv.Atoi(parts[2])
	if err != nil {
		sendError(w, http.StatusBadRequest, "ID должен быть числом")
		return
	}

	// Проверяем, что задача существует и принадлежит пользователю
	var exists bool
	err = config.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ? AND userid = ?)",
		taskID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		sendError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	// Читаем данные для обновления
	var requestData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	// Формируем запрос на обновление
	updates := []string{}
	params := []interface{}{}

	if requestData.Title != "" {
		updates = append(updates, "title = ?")
		params = append(params, strings.TrimSpace(requestData.Title))
	}
	if requestData.Description != "" {
		updates = append(updates, "description = ?")
		params = append(params, strings.TrimSpace(requestData.Description))
	}
	if requestData.Status != "" {
		validStatuses := []string{"pending", "in_progress", "completed"}
		isValid := false
		for _, s := range validStatuses {
			if requestData.Status == s {
				isValid = true
				break
			}
		}
		if isValid {
			updates = append(updates, "status = ?")
			params = append(params, requestData.Status)
		}
	}

	if len(updates) == 0 {
		sendError(w, http.StatusBadRequest, "Нет данных для обновления")
		return
	}

	// Добавляем условия WHERE
	params = append(params, taskID, userID)

	sql := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ? AND userid = ?", strings.Join(updates, ", "))
	_, err = config.DB.Exec(sql, params...)
	if err != nil {
		log.Printf("Ошибка обновления в БД: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось обновить задачу")
		return
	}

	// Получаем обновлённую задачу
	var task Task
	var createdAt string
	err = config.DB.QueryRow(
		"SELECT id, title, description, status, userid, created_at FROM tasks WHERE id = ? AND userid = ?",
		taskID, userID,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.UserID, &createdAt)

	if err != nil {
		log.Printf("Ошибка получения обновлённой задачи: %v", err)
		sendError(w, http.StatusInternalServerError, "Задача обновлена, но не удалось получить данные")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// Обработчик для маршрута DELETE /tasks/:id (удалить только свою задачу)
func taskDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	// Извлекаем ID из пути
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		sendError(w, http.StatusBadRequest, "Неверный формат пути")
		return
	}

	taskID, err := strconv.Atoi(parts[2])
	if err != nil {
		sendError(w, http.StatusBadRequest, "ID должен быть числом")
		return
	}

	// Удаляем задачу (только если она принадлежит пользователю)
	result, err := config.DB.Exec(
		"DELETE FROM tasks WHERE id = ? AND userid = ?",
		taskID, userID,
	)
	if err != nil {
		log.Printf("Ошибка удаления из БД: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось удалить задачу")
		return
	}

	// Проверяем, была ли удалена хотя бы одна строка
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка проверки удалённых строк: %v", err)
		sendError(w, http.StatusInternalServerError, "Не удалось проверить результат удаления")
		return
	}

	if rowsAffected == 0 {
		sendError(w, http.StatusNotFound, "Задача не найдена")
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "Задача удалена"})
}

// Главная функция для обработки всех запросов к /tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Если путь содержит ID (например, /tasks/123)
	if strings.HasPrefix(path, "/tasks/") && len(path) > len("/tasks/") {
		// Определяем метод и вызываем нужный обработчик
		switch r.Method {
		case http.MethodGet:
			taskGetByIDHandler(w, r)
		case http.MethodPut:
			taskPutHandler(w, r)
		case http.MethodDelete:
			taskDeleteHandler(w, r)
		default:
			sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		}
		return
	}

	// Если путь просто /tasks
	switch r.Method {
	case http.MethodGet:
		tasksGetHandler(w, r)
	case http.MethodPost:
		tasksPostHandler(w, r)
	default:
		sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
	}
}

// Вспомогательная функция для получения userID из запроса
func getUserIDFromRequest(r *http.Request) (int, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf("userID не найден")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// Вспомогательная функция для отправки JSON ответа
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Вспомогательная функция для отправки ошибки
func sendError(w http.ResponseWriter, statusCode int, message string) {
	sendJSON(w, statusCode, ErrorResponse{Error: message})
}

func main() {
	// Инициализируем логирование
	utils.InitLogger()

	// Загружаем конфигурацию
	if err := config.Load(); err != nil {
		utils.LogError(err, "Ошибка загрузки конфигурации")
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	utils.LogInfo("Запуск сервера", "port", config.Port)

	// Инициализируем базу данных
	if err := config.InitDB(); err != nil {
		utils.LogError(err, "Ошибка инициализации БД")
		log.Fatal("Ошибка инициализации БД:", err)
	}
	defer config.CloseDB()

	tasksHandlerNew := handlers.NewTasksHandler()

	// Используем порт из конфигурации
	port := config.Port

	// Используем разрешённые origins из конфигурации
	allowedOrigins := config.AllowedOrigins

	// Регистрируем маршруты с CORS middleware (API маршруты регистрируются первыми)
	http.HandleFunc("/info", middleware.CORS(allowedOrigins)(infoHandler))
	http.HandleFunc("/auth/register", middleware.CORS(allowedOrigins)(registerHandler))
	http.HandleFunc("/auth/login", middleware.CORS(allowedOrigins)(loginHandler))
	http.HandleFunc("/auth/reset-simple", middleware.CORS(allowedOrigins)(resetPasswordHandler))

	// Маршрут /me с разными методами
	http.HandleFunc("/me", middleware.CORS(allowedOrigins)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return // CORS уже обработал
		}
		switch r.Method {
		case http.MethodGet:
			middleware.Authenticate(meHandler)(w, r)
		case http.MethodPut:
			middleware.Authenticate(mePutHandler)(w, r)
		case http.MethodDelete:
			middleware.Authenticate(meDeleteHandler)(w, r)
		default:
			sendError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		}
	}))

	// Регистрируем маршруты для задач с использованием handlers
	http.HandleFunc("/tasks", middleware.CORS(allowedOrigins)(middleware.Authenticate(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tasksHandlerNew.GetTasks(w, r)
		case http.MethodPost:
			tasksHandlerNew.CreateTask(w, r)
		default:
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
		}
	})))

	// Маршрут /tasks/ для операций с конкретной задачей (GET, PUT, DELETE по ID)
	// Используем старую функцию tasksHandler для обработки запросов к /tasks/:id
	http.HandleFunc("/tasks/", middleware.CORS(allowedOrigins)(middleware.Authenticate(tasksHandler)))

	// Маршрут для загрузки файлов
	http.HandleFunc("/upload", handlers.UploadFileHandler)

	// Обработчик для статических файлов (фронтенд) - регистрируется последним
	// Получаем абсолютный путь к папке client
	workDir, _ := os.Getwd()
	staticDir := filepath.Join(workDir, "client")
	log.Printf("Статические файлы будут обслуживаться из: %s", staticDir)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			middleware.CORSHandler(allowedOrigins)(w, r)
			return
		}

		path := r.URL.Path

		// Если запрос к корню, отдаём index.html
		if path == "/" {
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
			http.NotFound(w, r)
			return
		}

		// Убираем начальный слэш для построения пути к файлу
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Безопасно строим путь к файлу
		filePath := filepath.Join(staticDir, cleanPath)

		// Проверяем, что путь находится внутри staticDir (защита от path traversal)
		relPath, err := filepath.Rel(staticDir, filePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			log.Printf("Попытка доступа вне staticDir: %s", path)
			http.NotFound(w, r)
			return
		}

		// Проверяем существование файла
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("Файл не найден: %s (запрошенный путь: %s)", filePath, path)
			http.NotFound(w, r)
			return
		}

		// Отдаём файл
		http.ServeFile(w, r, filePath)
	})

	// Создаём HTTP сервер
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: nil, // используем DefaultServeMux с зарегистрированными маршрутами
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на порту %s", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ждём сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Останавливаем сервер... ⛔")

	// Создаём контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем сервер, давая время активным запросам завершиться
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка остановки сервера: %v", err)
	}

	// Закрываем базу данных
	if err := config.CloseDB(); err != nil {
		log.Printf("Ошибка закрытия БД: %v", err)
	} else {
		log.Println("База закрыта. До встречи! 👋")
	}
}
