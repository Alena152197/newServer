package middleware

import (
	"net/http"
)

// CORS настраивает заголовки для междоменных запросов
func CORS(allowedOrigins []string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Проверяем, разрешён ли origin
			allowed := false
			if origin == "" {
				// Некоторые клиенты (curl, Postman) не отправляют Origin
				allowed = true
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Разрешаем методы
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			// Разрешаем заголовки
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Разрешаем отправку credentials (куки, авторизация)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Кэш preflight запросов (OPTIONS) на 10 минут
			w.Header().Set("Access-Control-Max-Age", "600")

			// Если это preflight запрос (OPTIONS), сразу отвечаем
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Вызываем следующий обработчик
			next(w, r)
		}
	}
}

// CORSHandler обрабатывает OPTIONS запросы
func CORSHandler(allowedOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := false
		if origin == "" {
			allowed = true
		} else {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "600")

		w.WriteHeader(http.StatusNoContent)
	}
}
