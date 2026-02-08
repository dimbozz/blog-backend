package middleware

import (
	"blog-backend/pkg/jwt"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет JWT токен и устанавливает контекст пользователя
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Получаем заголовок Authorization из запроса
		// Проверяем, что заголовок не пустой
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendAuthError(w, "Authorization header missing")
			return
		}

		// Проверяем формат "Bearer <token>"
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			sendAuthError(w, "Invalid authorization header format")
			return
		}

		// Извлекаем токен
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Валидируем токен с помощью ValidateToken() из auth.go
		claims, err := jwt.ValidateToken(tokenString)
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
