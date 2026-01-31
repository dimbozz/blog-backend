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

// CreateUserRequest для парсинга JSON
type CreateUserRequest struct {
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
	ID        int        `json:"id"`
	Title     string     `json:"title"`      // Заголовок поста
	Content   string     `json:"content"`    // Текст поста
	AuthorID  int        `json:"author_id"`  // ID автора комментария
	Status    string     `json:"status"`     // "draft" или "published"
	PublishAt *time.Time `json:"publish_at"` // через указатель, который может быть nil (для представления SQL NULL)
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
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

// DTO для создания поста (без ID, created_at, updated_at)
type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,max=255"`
	Content string `json:"content" validate:"required,max=5000"`
}

// DTO для обновления поста (опциональные поля)
type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,max=255"`
	Content *string `json:"content" validate:"omitempty,max=5000"`
}

// DTO для ответа (полная информация)
type PostResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DTO для списка постов
type ListPostsResponse struct {
	Posts []*PostResponse `json:"posts"`
	Total int             `json:"total"`
}
