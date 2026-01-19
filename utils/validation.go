package utils

import (
	"strings"
	"unicode"
)

// ValidateEmail проверяет, что email имеет правильный формат
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	return true
}

// ValidatePassword проверяет, что пароль соответствует требованиям
func ValidatePassword(password string) (bool, string) {
	password = strings.TrimSpace(password)
	if len(password) < 6 {
		return false, "Пароль должен быть не короче 6 символов"
	}
	if len(password) > 100 {
		return false, "Пароль слишком длинный"
	}
	return true, ""
}

// ValidateUsername проверяет имя пользователя
func ValidateUsername(username string) (bool, string) {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return false, "Имя пользователя должно быть не короче 3 символов"
	}
	if len(username) > 50 {
		return false, "Имя пользователя слишком длинное"
	}
	// Проверяем, что содержит только буквы, цифры и подчёркивания
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false, "Имя пользователя может содержать только буквы, цифры и подчёркивания"
		}
	}
	return true, ""
}

// ValidateTaskTitle проверяет заголовок задачи
func ValidateTaskTitle(title string) (bool, string) {
	title = strings.TrimSpace(title)
	if len(title) == 0 {
		return false, "Заголовок не может быть пустым"
	}
	if len(title) > 200 {
		return false, "Заголовок слишком длинный"
	}
	return true, ""
}

// ValidateTaskStatus проверяет статус задачи
func ValidateTaskStatus(status string) bool {
	validStatuses := []string{"pending", "in_progress", "completed"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}