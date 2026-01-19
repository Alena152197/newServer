// функции для авторизации и аутентификации
package utils

import (
	"time"

	"server_new/config"

	"github.com/golang-jwt/jwt/v5" // для работы с JWT токенами
	"golang.org/x/crypto/bcrypt"   // для хеширования паролей
)

// GenerateToken создаёт JWT токен для пользователя
func GenerateToken(userID int) (string, error) {
	// Создаём claims (данные в токене)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // токен действует 7 дней
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Создаём токен с алгоритмом HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken проверяет токен и возвращает UserID
func ValidateToken(tokenString string) (int, error) {
	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(JWT_SECRET), nil
	})

	if err != nil {
		return 0, err
	}

	// Извлекаем claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, jwt.ErrSignatureInvalid
}

// getJWTSecret получает секрет из конфигурации или использует значение по умолчанию
func getJWTSecret() string {
	if config.JWTSecret != "" {
		return config.JWTSecret
	}
	return "dev-secret-change-me" // только для разработки!
}

// JWT_SECRET — секретный ключ для подписи токенов
var JWT_SECRET = getJWTSecret()

// Claims — структура для данных в JWT токене
type Claims struct {
	UserID               int `json:"userId"`
	jwt.RegisteredClaims     // встроенная структура для стандартных полей
}

// HashPassword создаёт хэш пароля
func HashPassword(password string) (string, error) {
	// bcrypt.GenerateFromPassword создаёт хэш с "стоимостью" 10
	// Чем больше стоимость, тем безопаснее, но медленнее
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword проверяет, соответствует ли пароль хэшу
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
