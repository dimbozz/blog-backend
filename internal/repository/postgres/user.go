package postgres

import (
	"blog-backend/internal/models"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// CreateUser создает нового пользователя в базе данных
func CreateUser(email, username, passwordHash string) (*models.User, error) {
	// TODO: Реализуйте создание пользователя
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос для защиты от SQL-инъекций!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос с плейсхолдерами $1, $2, $3
	//    INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at
	// 2. Выполните запрос с db.QueryRow(query, email, username, passwordHash)
	// 3. Считайте результат в переменные user.ID и user.CreatedAt
	// 4. Заполните остальные поля структуры User
	// 5. Обработайте ошибки
	//
	// НИКОГДА не используйте fmt.Sprintf для построения SQL запросов!

	// 1. Создаем SQL запрос с плейсхолдерами $1, $2, $3
	query := `
        INSERT INTO users (email, username, password_hash) 
        VALUES ($1, $2, $3) 
        RETURNING id, created_at
    `
	// Инициализируем структуру User
	user := &models.User{}
	// 2. Выполняем запрос с db.QueryRow(query, email, username, passwordHash)
	// 3. Считываем результат в переменные user.ID и user.CreatedAt
	err := Db.QueryRow(query, email, username, passwordHash).Scan(&user.ID, &user.CreatedAt)
	// 5. Обрабатываем ошибки
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 4. Заполняем остальные поля структуры User
	user.Email = email
	user.Username = username

	return user, nil
}

// GetUserByEmail находит пользователя по email
func GetUserByEmail(email string) (*models.User, error) {
	// TODO: Реализуйте поиск пользователя по email
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос с плейсхолдером $1
	//    SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1
	// 2. Выполните запрос с db.QueryRow(query, email)
	// 3. Считайте все поля в структуру User с помощью Scan()
	// 4. Обработайте случай sql.ErrNoRows (пользователь не найден)
	//
	// Подсказка: используйте sql.ErrNoRows для проверки отсутствия результата

	// 1. Создаем SQL запрос с плейсхолдером $1
	query := `
        SELECT id, email, username, password_hash, created_at 
        FROM users 
        WHERE email = $1
    `
	// Инициализируем структуру User
	user := &models.User{}

	// 2. Выполняем запрос с db.QueryRow(query, email)
	// 3. Считываем все поля в структуру User с помощью Scan()
	err := Db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		// 4. Обрабатываем случай sql.ErrNoRows (пользователь не найден)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByID находит пользователя по ID
func GetUserByID(userID int) (*models.User, error) {
	// TODO: Реализуйте поиск пользователя по ID
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Создайте SQL запрос для поиска по ID
	// 2. НЕ включайте password_hash в SELECT (он не нужен для профиля)
	// 3. Выполните запрос и обработайте результат
	//
	// Запрос: SELECT id, email, username, created_at FROM users WHERE id = $1

	// 1. Создаем SQL запрос для поиска по ID
	query := `
        SELECT id, email, username, created_at 
        FROM users 
        WHERE id = $1
    `

	user := &models.User{}
	// 3. Выполняем запрос
	err := Db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
	)

	// Обрабатываем ошибку
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, fmt.Errorf("failed to get user by ID %d: %w", userID, err)
	}

	return user, nil
}

// UserExistsByEmail проверяет, существует ли пользователь с данным email
func UserExistsByEmail(email string) (bool, error) {
	// TODO: Реализуйте проверку существования пользователя
	// КРИТИЧЕСКИ ВАЖНО: Используйте параметризованный запрос!
	//
	// Что нужно сделать:
	// 1. Используйте SQL функцию EXISTS для эффективной проверки
	//    SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	// 2. Результат будет булевым значением
	// 3. Считайте результат в переменную типа bool
	//
	// Это эффективнее чем получать полную запись пользователя

	query := `
        SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
    `

	var ifUserExists bool
	err := Db.QueryRow(query, email).Scan(&ifUserExists)

	if err != nil {
		return false, fmt.Errorf("failed to check user exists: %w", err)
	}

	return ifUserExists, nil
}
