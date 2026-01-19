package middleware

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse структура для ошибки
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// AppError представляет ошибку приложения
type AppError struct {
	Message    string
	StatusCode int
	Details    string
}

func (e *AppError) Error() string {
	return e.Message
}

// ErrorHandler обрабатывает ошибки и отправляет JSON ответ
func ErrorHandler(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Ошибка: %v", err)

	response := ErrorResponse{
		Error: err.Error(),
		Code:  statusCode,
	}

	// Если это AppError, используем его данные
	if appErr, ok := err.(*AppError); ok {
		response.Error = appErr.Message
		response.Code = appErr.StatusCode
		if appErr.Details != "" {
			response.Details = appErr.Details
		}
		statusCode = appErr.StatusCode
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SendError отправляет ошибку клиенту
func SendError(w http.ResponseWriter, message string, statusCode int) {
	ErrorHandler(w, &AppError{
		Message:    message,
		StatusCode: statusCode,
	}, statusCode)
}

// SendValidationError отправляет ошибку валидации
func SendValidationError(w http.ResponseWriter, message string) {
	SendError(w, message, http.StatusBadRequest)
}

// SendNotFound отправляет ошибку "не найдено"
func SendNotFound(w http.ResponseWriter, message string) {
	SendError(w, message, http.StatusNotFound)
}

// SendUnauthorized отправляет ошибку авторизации
func SendUnauthorized(w http.ResponseWriter, message string) {
	SendError(w, message, http.StatusUnauthorized)
}

// SendInternalError отправляет внутреннюю ошибку сервера
func SendInternalError(w http.ResponseWriter, err error) {
	log.Printf("Внутренняя ошибка: %v", err)
	SendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
}