// service/mock_user_repo_test.go
package service_test

import (
	"context"
	"fmt"

	"blog-backend/internal/model"
	"blog-backend/internal/repository"
)

type MockUserRepo struct {
	users map[int]*model.User
}

func NewMockUserRepo() repository.UserRepository {
	return &MockUserRepo{
		users: map[int]*model.User{
			1: {
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
	}
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %d", id)
	}
	return user, nil
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found by email: %s", email)
}

func (m *MockUserRepo) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, user := range m.users {
		if user.Email == email {
			return true, nil // Пользователь существует
		}
	}
	return false, nil // Пользователь не найден
}

func (m *MockUserRepo) CreateUser(ctx context.Context, name, email, password string) (*model.User, error) {
	user := &model.User{
		ID:           len(m.users) + 1,
		Username:     name,
		Email:        email,
		PasswordHash: password,
	}
	m.users[user.ID] = user
	return user, nil
}

func (m *MockUserRepo) Update(ctx context.Context, user *model.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %d", user.ID)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) Delete(ctx context.Context, id int) error {
	if _, exists := m.users[id]; !exists {
		return fmt.Errorf("user not found: %d", id)
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	var result []*model.User
	for _, user := range m.users {
		result = append(result, user)
	}
	return result, len(result), nil
}

func (m *MockUserRepo) Count(ctx context.Context) (int, error) {
	return len(m.users), nil
}

// Компиляторная проверка показывает недостающие методы
var _ repository.UserRepository = (*MockUserRepo)(nil)
