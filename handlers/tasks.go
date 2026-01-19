package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"server_new/services"
	"server_new/utils"
)

type TasksHandler struct {
	service *services.TasksService
}

func NewTasksHandler() *TasksHandler {
	return &TasksHandler{
		service: services.NewTasksService(),
	}
}

// Вспомогательные функции
func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	sendJSON(w, statusCode, map[string]string{"error": message})
}

// getUserID извлекает userID из заголовка запроса
func getUserID(r *http.Request) (int, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf("userID не найден")
	}
	return strconv.Atoi(userIDStr)
}

// GetTasks получение списка задач
// @Summary Получить список задач
// @Description Возвращает список задач текущего пользователя с пагинацией
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество на странице" default(10)
// @Param status query string false "Фильтр по статусу" Enums(pending, in_progress, completed)
// @Success 200 {array} models.Task
// @Header 200 {string} X-Total-Count "Общее количество задач"
// @Failure 401 {object} map[string]string
// @Router /tasks [get]
// @Security BearerAuth
func (h *TasksHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	status := r.URL.Query().Get("status")

	tasks, total, err := h.service.GetTasksByUserID(userID, page, limit, status)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Не удалось получить задачи")
		return
	}

	// Устанавливаем заголовок с общим количеством задач
	w.Header().Set("X-Total-Count", strconv.Itoa(total))

	sendJSON(w, http.StatusOK, tasks)
}

func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Не удалось определить пользователя")
		return
	}

	var requestData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	title := strings.TrimSpace(requestData.Title)
	if title == "" {
		sendError(w, http.StatusBadRequest, "Укажи непустой title")
		return
	}

	description := strings.TrimSpace(requestData.Description)
	status := strings.TrimSpace(requestData.Status)
	if status == "" {
		status = "pending"
	}

	// Логируем начало операции
	utils.LogInfo("Создание задачи", "userID", userID, "title", title)

	task, err := h.service.CreateTask(title, description, status, userID)
	if err != nil {
		// Логируем ошибку
		utils.LogError(err, "Ошибка создания задачи", "userID", userID)
		sendError(w, http.StatusInternalServerError, "Не удалось создать задачу")
		return
	}

	// Логируем успех
	utils.LogInfo("Задача создана", "taskID", task.ID, "userID", userID)

	sendJSON(w, http.StatusCreated, task)
}
