package middleware

import (
	"blog-backend/pkg/jwt"
	"context"
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
			AbortError(w, r, "Authorization header missing", http.StatusUnauthorized, nil)
			return
		}

		// Проверяем формат "Bearer <token>"
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			AbortError(w, r, "Invalid authorization header format", http.StatusUnauthorized, nil)
			return
		}

		// Извлекаем токен
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Валидируем токен с помощью ValidateToken() из auth.go
		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			AbortError(w, r, "Invalid token", http.StatusUnauthorized, err)
			return
		}

		// 6. Добавьте данные пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)

		// 7. Передаем управление следующему обработчику
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
