// Сервис для задач будет содержать всю логику работы с базой данных

package services

import (
	"database/sql"
	"fmt"
	"strings"

	"server_new/config"
	"server_new/models"
)

// TasksService содержит методы для работы с задачами
type TasksService struct {
	db *sql.DB
}

// NewTasksService создаёт новый экземпляр сервиса
func NewTasksService() *TasksService {
	return &TasksService{db: config.DB}
}

// GetTasksByUserID возвращает задачи пользователя с пагинацией и фильтрацией
func (s *TasksService) GetTasksByUserID(userID, page, limit int, status string) ([]models.Task, int, error) {
	offset := (page - 1) * limit

	query := "SELECT id, title, description, status, userid FROM tasks WHERE userid = ?"
	args := []interface{}{userID}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка запроса к БД: %v", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.UserID)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	// Получаем общее количество
	var total int
	countQuery := "SELECT COUNT(*) FROM tasks WHERE userid = ?"
	if status != "" {
		countQuery += " AND status = ?"
		err = s.db.QueryRow(countQuery, userID, status).Scan(&total)
	} else {
		err = s.db.QueryRow(countQuery, userID).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения количества: %v", err)
	}

	return tasks, total, nil
}

// CreateTask создаёт новую задачу
func (s *TasksService) CreateTask(title, description, status string, userID int) (*models.Task, error) {
	result, err := s.db.Exec(
		"INSERT INTO tasks (title, description, status, userid) VALUES (?, ?, ?, ?)",
		title, description, status, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка вставки в БД: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ID: %v", err)
	}

	return &models.Task{
		ID:          int(id),
		Title:       title,
		Description: description,
		Status:      status,
		UserID:      userID,
	}, nil
}

// GetTaskByID возвращает задачу по ID (только если она принадлежит пользователю)
func (s *TasksService) GetTaskByID(taskID, userID int) (*models.Task, error) {
	var task models.Task
	var createdAt string
	err := s.db.QueryRow(
		"SELECT id, title, description, status, userid, created_at FROM tasks WHERE id = ? AND userid = ?",
		taskID, userID,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.UserID, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("ошибка запроса к БД: %v", err)
	}

	return &task, nil
}

// UpdateTask обновляет задачу (только если она принадлежит пользователю)
func (s *TasksService) UpdateTask(taskID, userID int, title, description, status string) (*models.Task, error) {
	// Проверяем существование и владельца
	_, err := s.GetTaskByID(taskID, userID)
	if err != nil {
		return nil, err
	}

	updates := []string{}
	params := []interface{}{}

	if title != "" {
		updates = append(updates, "title = ?")
		params = append(params, title)
	}
	if description != "" {
		updates = append(updates, "description = ?")
		params = append(params, description)
	}
	if status != "" {
		updates = append(updates, "status = ?")
		params = append(params, status)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("нет данных для обновления")
	}

	params = append(params, taskID, userID)

	sql := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ? AND userid = ?", strings.Join(updates, ", "))
	_, err = s.db.Exec(sql, params...)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления: %v", err)
	}

	return s.GetTaskByID(taskID, userID)
}

// DeleteTask удаляет задачу (только если она принадлежит пользователю)
func (s *TasksService) DeleteTask(taskID, userID int) error {
	result, err := s.db.Exec(
		"DELETE FROM tasks WHERE id = ? AND userid = ?",
		taskID, userID,
	)
	if err != nil {
		return fmt.Errorf("ошибка удаления: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}

	return nil
}
