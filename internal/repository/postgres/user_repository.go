package postgres

import (
	"blog-backend/internal/model"
	"blog-backend/service"
	"context"
	"database/sql"
	"fmt"
)

type PostgresUserRepository struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:  db,
		ctx: context.Background(), // ← Контекст!
	}
}

func (r *PostgresUserRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, name, created_at FROM users WHERE id = $1`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Username, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, service.ErrUserNotFound
	}
	return &user, err
}

// CreateUser создает нового пользователя в базе данных
func (r *PostgresUserRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
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
	user := &model.User{}
	// 2. Выполняем запрос с db.QueryRow(query, email, username, passwordHash)
	// 3. Считываем результат в переменные user.ID и user.CreatedAt
	err := r.db.QueryRowContext(ctx, query, email, username, passwordHash).Scan(&user.ID, &user.CreatedAt)
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
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
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
	user := &model.User{}

	// 2. Выполняем запрос с db.QueryRow(query, email)
	// 3. Считываем все поля в структуру User с помощью Scan()
	err := r.db.QueryRowContext(ctx, query, email).Scan(
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
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
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

	user := &model.User{}
	// 3. Выполняем запрос
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
	)

	// Проверить !!!
	// if errors.Is(err, sql.ErrNoRows) {
	//     return nil, service.ErrUserNotFound
	// }

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
func (r *PostgresUserRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
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

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check user exists: %w", err)
	}

	return exists, nil
}
