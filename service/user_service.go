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

	user, err := s.userRepo.CreateUser(ctx, email, username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id int) (*model.User, error) {

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// UserExistsByEmail проверяет, существует ли пользователь с данным email
func (s *UserService) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.userRepo.UserExistsByEmail(ctx, email)
}

// GetUserByEmail возвращает пользователя по email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
