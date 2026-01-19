package models

// Task представляет задачу в базе данных
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserID      int    `json:"userid"` // ID пользователя, которому принадлежит задача
}
