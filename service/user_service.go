package service

import (
	"blog-backend/internal/model"
	"blog-backend/internal/repository"
	"blog-backend/pkg/jwt"
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

// Register создает нового пользователя и возвращает токен
func (s *UserService) Register(ctx context.Context, email, username, passwordHash string) (*model.User, string, error) {

	user, err := s.userRepo.CreateUser(ctx, email, username, passwordHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := jwt.GenerateToken(*user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
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

// Login выполняет авторизацию пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	// 1. Находим пользователя
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return nil, "", fmt.Errorf("Invalid email or password")
	}

	// 2. Проверяем пароль
	if !jwt.CheckPassword(password, user.PasswordHash) {
		return nil, "", fmt.Errorf("Invalid email or password")
	}

	// 3. Генерируем JWT токен
	token, err := jwt.GenerateToken(*user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}
