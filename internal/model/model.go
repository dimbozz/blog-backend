package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User представляет пользователя в системе
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // "-" исключает поле из JSON
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest структура для запроса регистрации
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest структура для запроса входа
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse структура ответа с токеном
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Claims структура для JWT токена
type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Post struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`     // Заголовок поста
	Content   string    `json:"content"`   // Текст поста
	AuthorID  int       `json:"author_id"` // ID автора комментария
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    string    `json:"post_id"`             // Связь с постом
	AuthorID  int       `json:"author_id"`           // ID автора комментария
	Content   string    `json:"content"`             // Текст комментария
	ParentID  *int      `json:"parent_id,omitempty"` // ID родительского комментария, nil = корневой комментарий)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
