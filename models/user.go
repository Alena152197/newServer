// структура для хранения данных пользователя
package models

import (
	"time"
)

// User представляет пользователя в базе данных
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // не включаем пароль в JSON ответы
	CreatedAt time.Time `json:"created_at"`
}

// UserResponse — данные пользователя для ответа (без пароля)
type UserResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}