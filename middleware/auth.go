// Middleware — это функции, которые выполняются перед основным обработчиком. Например, middleware для авторизации проверяет токен перед тем, как запрос дойдёт до обработчика
package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"server_new/utils"
)

// Authenticate проверяет JWT токен и добавляет userID в контекст запроса
func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, http.StatusUnauthorized, "Нужен токен авторизации")
			return
		}

		// Формат: "Bearer <токен>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendError(w, http.StatusUnauthorized, "Неверный формат токена")
			return
		}

		token := parts[1]

		// Проверяем токен
		userID, err := utils.ValidateToken(token)
		if err != nil {
			sendError(w, http.StatusUnauthorized, "Токен недействителен или истёк")
			return
		}

		// Сохраняем userID в заголовке запроса (временное решение)
		// В более продвинутых версиях можно использовать контекст
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

		// Вызываем следующий обработчик
		next(w, r)
	}
}

// sendError отправляет ошибку в формате JSON
func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
