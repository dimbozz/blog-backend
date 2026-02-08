// internal/handlers/auth_handler_test.go
package handlers_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-backend/internal/handlers"
	"blog-backend/internal/repository"
	"blog-backend/service"
)

// setupAuthTestRouter - создает полный тестовый роутер
func setupAuthTestRouter() (http.Handler, repository.UserRepository) {
	// Используем существующий UserService из user_service.go
	userRepo := NewMemoryUserRepository()
	// cfg := NewTestConfig()

	// UserService принимает userRepo
	userSvc := service.NewUserService(userRepo)
	logger := log.New(io.Discard, "", 0)

	userHandler := handlers.NewUserHandler(userSvc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", userHandler.RegisterHandler)
	mux.HandleFunc("POST /api/auth/login", userHandler.LoginHandler)

	return mux, userRepo
}

// TestRegisterHandler - Регистрация пользователей
func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupUser      bool
		expectedStatus int
	}{
		{
			name:           "valid_register",
			body:           `{"email": "user@example.com", "username": "user", "password": "password123"}`,
			setupUser:      false,
			expectedStatus: http.StatusCreated, // 201
		},
		{
			name:           "user_already_exists",
			body:           `{"email": "exists@example.com", "username": "exists", "password": "pass"}`,
			setupUser:      true,
			expectedStatus: http.StatusConflict, // 409
		},
		{
			name:           "invalid_json",
			body:           `{invalid json`,
			setupUser:      false,
			expectedStatus: http.StatusBadRequest, // 400
		},
		{
			name:           "missing_fields",
			body:           `{"email": ""}`,
			setupUser:      false,
			expectedStatus: http.StatusBadRequest, // 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, userRepo := setupAuthTestRouter()

			// Создаем существующего пользователя для теста конфликта
			if tt.setupUser {
				ctx := context.Background()
				userRepo.CreateUser(ctx, "exists@example.com", "exists", "hash")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
				bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				bodyBytes, _ := io.ReadAll(w.Body)
				t.Logf("Status: %d, Body: %s", w.Code, string(bodyBytes))
				t.Errorf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestLoginHandler - Вход пользователей
func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setupUser      bool
		expectedStatus int
	}{
		{
			name:           "valid_login",
			body:           `{"email": "test@example.com", "password": "password123"}`,
			setupUser:      true,
			expectedStatus: http.StatusOK, // 200 + JWT token
		},
		{
			name:           "user_not_found",
			body:           `{"email": "unknown@example.com", "password": "pass"}`,
			setupUser:      false,
			expectedStatus: http.StatusUnauthorized, // 401
		},
		{
			name:           "wrong_password",
			body:           `{"email": "test@example.com", "password": "wrongpass"}`,
			setupUser:      true,
			expectedStatus: http.StatusUnauthorized, // 401
		},
		{
			name:           "invalid_json",
			body:           `{invalid`,
			setupUser:      false,
			expectedStatus: http.StatusBadRequest, // 400
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, userRepo := setupAuthTestRouter()

			// Создаем тестового пользователя для валидного логина
			if tt.setupUser {
				ctx := context.Background()
				// Пароль "password123" → хеш для jwt.CheckPassword()
				userRepo.CreateUser(ctx, "test@example.com", "testuser", "password123_hash")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
				bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				bodyBytes, _ := io.ReadAll(w.Body)
				t.Logf("Status: %d, Body: %s", w.Code, string(bodyBytes))
				t.Errorf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
