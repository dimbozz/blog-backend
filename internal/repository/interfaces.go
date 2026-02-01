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
	// CRUD
	CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
	GetPostByID(ctx context.Context, id int) (*model.Post, error)
	UpdatePost(ctx context.Context, id int, post *model.Post) (*model.Post, error)
	DeletePost(ctx context.Context, id int) error

	// Список + пагинация
	ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error)
	CountPosts(ctx context.Context) (int, error)

	// Методы планировщика
	GetReadyToPublish(ctx context.Context, batchSize int) ([]*model.Post, error)
    PublishPost(ctx context.Context, postID int) error
}

// UserRepository — интерфейс для работы с пользователями
type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, email string, username string, passwordHash string) (*model.User, error)
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
}
