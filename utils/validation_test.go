package utils

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"валидный email", "test@example.com", true},
		{"email без @", "testexample.com", false},
		{"пустой email", "", false},
		{"только домен", "@example.com", false},
		{"только имя", "test@", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateEmail(tt.email); got != tt.want {
				t.Errorf("ValidateEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
		wantMsg  string
	}{
		{"короткий пароль", "12345", false, "Пароль должен быть не короче 6 символов"},
		{"валидный пароль", "password123", true, ""},
		{"длинный пароль", strings.Repeat("a", 101), false, "Пароль слишком длинный"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, msg := ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
			if msg != tt.wantMsg {
				t.Errorf("ValidatePassword() msg = %v, want %v", msg, tt.wantMsg)
			}
		})
	}
}


func TestValidateTaskTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"пустой заголовок", "", false},
		{"валидный заголовок", "Купить молоко", true},
		{"слишком длинный", strings.Repeat("a", 201), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ValidateTaskTitle(tt.title)
			if got != tt.want {
				t.Errorf("ValidateTaskTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

