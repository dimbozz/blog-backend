package handlers

import (
	"blog-backend/internal/model"
	"blog-backend/internal/repository/postgres"
	"blog-backend/pkg/auth"
	"blog-backend/pkg/jwt"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// RegisterHandler обрабатывает регистрацию нового пользователя
func RegisterHandler(userRepo *postgres.PostgresUserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()

		// TODO: Реализуйте регистрацию пользователя
		//
		// Пошаговый план:
		// 1. Распарсите JSON из тела запроса в структуру RegisterRequest
		// 2. Проведите валидацию данных (email, username, password)
		// 3. Проверьте, что пользователь с таким email не существует
		// 4. Захешируйте пароль с помощью функции HashPassword()
		// 5. Создайте пользователя в БД с помощью CreateUser()
		// 6. Сгенерируйте JWT токен с помощью GenerateToken()
		// 7. Верните ответ с токеном и данными пользователя
		//
		// Подсказки:
		// - Используйте json.NewDecoder(r.Body).Decode() для парсинга JSON
		// - Проверьте что все обязательные поля заполнены
		// - При ошибках возвращайте соответствующие HTTP статусы
		// - 400 для невалидных данных, 409 для дубликатов, 500 для внутренних ошибок
		// - Не забудьте установить Content-Type: application/json для ответа

		// 1. Парсим JSON
		var req model.RegisterRequest
		if err := parseJSONRequest(r, &req); err != nil {
			sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 2. Валидация
		if err := validateRegisterRequest(&req); err != nil {
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 3. Проверяем существование email
		if exists, err := userRepo.UserExistsByEmail(ctx, req.Email); err != nil {
			log.Printf("Database error: %v", err)
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if exists {
			sendErrorResponse(w, "User with this email already exists", http.StatusConflict)
			return
		}

		// 4. Хешируем пароль
		passwordHash, err := jwt.HashPassword(req.Password)
		if err != nil {
			log.Printf("Hash password error: %v", err)
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 5. Создаем пользователя
		user, err := userRepo.CreateUser(ctx, req.Email, req.Username, passwordHash)
		if err != nil {
			log.Printf("Create user error: %v", err)
			sendErrorResponse(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// 6. Генерируем токен
		token, err := jwt.GenerateToken(*user)
		if err != nil {
			log.Printf("Generate token error: %v", err)
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 7. Успешный ответ
		response := map[string]interface{}{
			"message": "User registered successfully",
			"user": map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"username": user.Username,
			},
			"token": token,
		}
		sendJSONResponse(w, response, http.StatusCreated)
	}
}

// LoginHandler обрабатывает вход пользователя
func LoginHandler(userRepo *postgres.PostgresUserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		// TODO: Реализуйте авторизацию пользователя
		//
		// Пошаговый план:
		// 1. Распарсите JSON из тела запроса в структуру LoginRequest
		// 2. Проведите базовую валидацию (email и password не пустые)
		// 3. Найдите пользователя по email с помощью GetUserByEmail()
		// 4. Проверьте пароль с помощью CheckPassword()
		// 5. Сгенерируйте JWT токен с помощью GenerateToken()
		// 6. Верните ответ с токеном и данными пользователя
		//
		// Важные моменты безопасности:
		// - При неверном email или пароле возвращайте одинаковое сообщение
		//   "Invalid email or password" чтобы не раскрывать существование email
		// - Используйте HTTP статус 401 для неверных учетных данных
		// - Не возвращайте password_hash в ответе

		// 1. Парсим JSON
		var req model.LoginRequest
		if err := parseJSONRequest(r, &req); err != nil {
			sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 2. Валидация
		if err := validateLoginRequest(&req); err != nil {
			sendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 3. Находим пользователя
		user, err := userRepo.GetUserByEmail(ctx, req.Email)
		if err != nil {
			log.Printf("Database error: %v", err)
			sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		if user == nil {
			sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// 4. Проверяем пароль
		if !jwt.CheckPassword(req.Password, user.PasswordHash) {
			sendErrorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// 5. Генерируем токен
		token, err := jwt.GenerateToken(*user)
		if err != nil {
			log.Printf("Generate token error: %v", err)
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 6. Успешный ответ
		response := map[string]interface{}{
			"message": "Login successful",
			"user": map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"username": user.Username,
			},
			"token": token,
		}
		sendJSONResponse(w, response, http.StatusOK)
	}
}

// ProfileHandler возвращает профиль текущего пользователя
// Этот обработчик вызывается только после AuthMiddleware
func ProfileHandler(userRepo *postgres.PostgresUserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()

		// Получаем userID из контекста
		// Контекст уже должен содержать userID
		userID, ok := auth.GetUserIDFromContext(r)
		if !ok {
			sendErrorResponse(w, "User ID not found in context", http.StatusInternalServerError)
			return
		}

		// Загружаем данные пользователя из БД с помощью GetUserByID()
		user, err := userRepo.GetUserByID(ctx, userID)
		if err != nil {
			log.Printf("Database error: %v", err)
			sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Если пользователь не найден - возвращаем 404
		if user == nil {
			sendErrorResponse(w, "User not found", http.StatusNotFound)
			return
		}

		// Отправляем профиль (без password_hash)
		response := map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"created_at": user.CreatedAt,
		}
		// Возвращаем данные пользователя в JSON формате
		sendJSONResponse(w, response, http.StatusOK)
	}
}

// sendJSONResponse отправляет JSON ответ (вспомогательная функция)
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendErrorResponse отправляет JSON ответ с ошибкой (вспомогательная функция)
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// parseJSONRequest парсит JSON из тела запроса (вспомогательная функция)
func parseJSONRequest(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Строгая проверка полей

	return decoder.Decode(v)
}

// validateRegisterRequest валидирует данные регистрации
func validateRegisterRequest(req *model.RegisterRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// TODO: Добавить дополнительные проверки
	// - Используйте ValidateEmail() и ValidatePassword() из auth.go
	// - Проверьте длину username (например, минимум 3 символа)
	// - Проверьте что username содержит только допустимые символы

	return nil
}

// validateLoginRequest валидирует данные входа
func validateLoginRequest(req *model.LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
