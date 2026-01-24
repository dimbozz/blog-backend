package service

import (
	"blog-backend/internal/model"
	"blog-backend/internal/repository"
	"context"
	"fmt"
)

type UserService struct {
	userRepo repository.UserRepository // интерфейс для гибкости
}

func NewUserService(ur repository.UserRepository) *UserService {
	return &UserService{
		userRepo: ur,
	}
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	user := &model.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser возвращает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, id int) (*model.User, error) {

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
