package auth

import "net/http"

type contextKey string

const (
	contextKeyUser = contextKey("user")
)

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(r *http.Request) (int, bool) {
	// Используем r.Context().Value("userID")
	// Проводим type assertion к int
	userID, ok := r.Context().Value("userID").(int)

	// Возвращаем значение и булевый флаг успешности
	return userID, ok
}
