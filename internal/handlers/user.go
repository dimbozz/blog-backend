package handlers

import (
	"blog-backend/internal/handlers/middleware"
	"blog-backend/internal/model"
	"blog-backend/pkg/auth"
	"blog-backend/pkg/jwt"
	"blog-backend/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// UserHandler обрабатывает HTTP запросы для пользователей
type UserHandler struct {
	userService *service.UserService
	log         *log.Logger
}

// NewUserHandler создает новый UserHandler
func NewUserHandler(userService *service.UserService, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         logger,
	}
}

// RegisterHandler обрабатывает регистрацию нового пользователя
func (h *UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Только POST
	if r.Method != http.MethodPost {
		middleware.AbortError(w, r, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	ctx := r.Context()

	// Реализуем регистрацию пользователя
	//
	// Пошаговый план:
	// 1. Распарсить JSON из тела запроса в структуру RegisterRequest
	// 2. Провести валидацию данных (email, username, password)
	// 3. Проверить, что пользователь с таким email не существует
	// 4. Захешировать пароль с помощью функции HashPassword()
	// 5. Создать пользователя в БД с помощью CreateUser()
	// 6. Сгенеририровать JWT токен с помощью GenerateToken()
	// 7. Вернуть ответ с токеном и данными пользователя
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
		middleware.AbortError(w, r, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	// 2. Валидация
	if err := validateRegisterRequest(&req); err != nil {
		middleware.AbortError(w, r, err.Error(), http.StatusBadRequest, err)
		return
	}

	// 3. Проверяем существование email
	if exists, err := h.userService.UserExistsByEmail(ctx, req.Email); err != nil {
		middleware.AbortError(w, r, "Database error", http.StatusInternalServerError, err)
		return
	} else if exists {
		middleware.AbortError(w, r, "Email already exists", http.StatusConflict, nil)
		return
	}

	// 4. Хешируем пароль
	passwordHash, err := jwt.HashPassword(req.Password)
	if err != nil {
		middleware.AbortError(w, r, "Failed to hash password", http.StatusInternalServerError, err)
		return
	}

	// 5. Создаем пользователя и токен
	user, token, err := h.userService.Register(ctx, req.Email, req.Username, passwordHash)
	if err != nil {
		middleware.AbortError(w, r, "Failed to create user", http.StatusInternalServerError, err)
		return
	}

	// 7. Возвращаем ответ с токеном и данными пользователя
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

// LoginHandler обрабатывает вход пользователя
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		middleware.AbortError(w, r, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}

	ctx := r.Context()
	// Авторизация пользователя
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
		middleware.AbortError(w, r, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	// 2. Валидация
	if err := validateLoginRequest(&req); err != nil {
		middleware.AbortError(w, r, "Invalid email or password", http.StatusBadRequest, err)
		return
	}

	// 3. Вызываем сервис
	user, token, err := h.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		middleware.AbortError(w, r, "Invalid email or password", http.StatusUnauthorized, err)
		return
	}

	// 4. Успешный ответ
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

// ProfileHandler возвращает профиль текущего пользователя
// Этот обработчик должен быть вызван только после AuthMiddleware
func (h *UserHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Получаем userID из контекста
	// Контекст уже должен содержать userID
	userID, ok := auth.GetUserIDFromContext(r)
	if !ok {
		middleware.AbortError(w, r, "User ID not found in context", http.StatusInternalServerError, nil)
		return
	}

	// Загружаем данные пользователя из БД с помощью GetUserByID()
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		middleware.AbortError(w, r, "Database error", http.StatusInternalServerError, err)
		return
	}
	// Если пользователь не найден - возвращаем 404
	if user == nil {
		middleware.AbortError(w, r, "User not found", http.StatusNotFound, nil)
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

// sendJSONResponse отправляет JSON ответ (вспомогательная функция)
func sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
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
