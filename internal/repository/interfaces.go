package repository

import (
	"context"

	"blog-backend/internal/model"
)

// Отдельный интерфейс для health checks
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// PostRepository — интерфейс для работы с постами
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id int) (*model.Post, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id int) error
}

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	FindByID(ctx context.Context, id int) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}
