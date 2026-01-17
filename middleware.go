package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	// 1. Импортируйте "context" и "strings"
	"context"
	"strings"
)

type contextKey string

const (
	contextKeyUser = contextKey("user")
)

// AuthMiddleware проверяет JWT токен и устанавливает контекст пользователя
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Реализуйте проверку JWT токена
		// Используйте:
		// - r.Header.Get("Authorization")
		// - strings.TrimPrefix(authHeader, "Bearer ")
		// - context.WithValue(r.Context(), "userID", claims.UserID)
		// - next.ServeHTTP(w, r.WithContext(ctx))

		// 2. Получаем заголовок Authorization из запроса
		// 3. Проверяем, что заголовок не пустой
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendAuthError(w, "Authorization header missing")
			return
		}

		// 4. Проверяем формат "Bearer <token>"
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			sendAuthError(w, "Invalid authorization header format")
			return
		}

		// 4. Извлекаем токен
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 5. Валидируем токен с помощью ValidateToken() из auth.go
		claims, err := ValidateToken(tokenString)
		if err != nil {
			sendAuthError(w, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// 6. Добавьте данные пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)

		// 7. Передаем управление следующему обработчику
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// sendAuthError отправляет JSON ответ с ошибкой 401 Unauthorized
func sendAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	_ = json.NewEncoder(w).Encode(map[string]string{
		// Если токен невалиден - возвращаем 401 Unauthorized
		// Если токен отсутствует - возвращаем 401 Unauthorized
		"error":   "401 Unauthorized",
		"message": message,
	})
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(r *http.Request) (int, bool) {
	// TODO: Реализуйте извлечение userID из контекста
	//
	// Что нужно сделать:
	// 1. Используйте r.Context().Value("userID")
	// 2. Проведите type assertion к int
	// 3. Верните значение и булевый флаг успешности
	//
	// Пример: userID, ok := r.Context().Value("userID").(int)
	userID, ok := r.Context().Value("userID").(int)

	// Возвращаем значение и булевый флаг успешности
	return userID, ok
}
