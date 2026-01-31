package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"blog-backend/internal/model"
)

// Реализация PostRepository для PostgreSQL (с CRUD методами)
type PostgresPostRepository struct {
	db *sql.DB
}

// Создаем новый репозиторий постов с подключением к БД
func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

// Создаем пост в БД и возвращает готовую запись с ID и временными метками
func (r *PostgresPostRepository) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	// INSERT с RETURNING возвращает все поля созданной записи
	query := `
        INSERT INTO posts (author_id, title, content, status, publish_at) 
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id, author_id, title, content, status, publish_at, created_at, updated_at`

	// Инициализируем структуру createdPost
	createdPost := &model.Post{}
	var publishAtNull sql.NullTime  // Для чтения из БД

	post.Status = "published"
	now := time.Now()

	if post.PublishAt != nil && post.PublishAt.After(now) {
		post.Status = "draft"
	}

	// sql.NullTime для БД
	var publishAtParam sql.NullTime
	if post.PublishAt != nil {
		publishAtParam.Time = *post.PublishAt
		publishAtParam.Valid = true
	}

	// Выполняем INSERT, передаем только данные (author_id, title, content, publish_at)
	row := r.db.QueryRowContext(
		ctx,
		query,
		post.AuthorID,
		post.Title,
		post.Content,
		post.Status,
		publishAtParam,
	)

	// БД заполняет все поля (ID генерируется автоматически)
	err := row.Scan(
		&createdPost.ID,        // Автогенерированный ID
		&createdPost.AuthorID,  // Из параметров INSERT
		&createdPost.Title,     // Из параметров INSERT
		&createdPost.Content,   // Из параметров INSERT
		&createdPost.Status,    // Из параметров INSERT
		&publishAtNull, // CURRENT_TIMESTAMP
		&createdPost.CreatedAt, // CURRENT_TIMESTAMP
		&createdPost.UpdatedAt, // CURRENT_TIMESTAMP
	)

	// ✅ NULL → *time.Time
    if publishAtNull.Valid {
        createdPost.PublishAt = &publishAtNull.Time
    }

	// Обрабатываем ошибки
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return createdPost, nil
}

// Получаем пост по ID
func (r *PostgresPostRepository) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	// Инициализируем структуру post
	post := &model.Post{}

	// SELECT одной записи по первичному ключу
	query := `
        SELECT id, author_id, title, content, created_at, updated_at 
        FROM posts 
        WHERE id = $1`

	// Выполняем SELECT
	row := r.db.QueryRowContext(ctx, query, id)

	// Заполняем структуру данными из БД
	err := row.Scan(
		&post.ID,
		&post.AuthorID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	// Обрабатываем ошибки
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return post, nil
}

// Обновляем пост и возвращает актуальную версию с updated_at
func (r *PostgresPostRepository) UpdatePost(ctx context.Context, id int, post *model.Post) (*model.Post, error) {
	// UPDATE с автоматическим updated_at и RETURNING всех полей
	query := `
        UPDATE posts 
        SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, author_id, title, content, created_at, updated_at`

	// Инициализируем структуру post
	updatedPost := &model.Post{}

	// Выполняем UPDATE
	row := r.db.QueryRowContext(ctx, query, post.Title, post.Content, id)

	// Заполняем структуру данными из БД
	err := row.Scan(
		&updatedPost.ID,
		&updatedPost.AuthorID,
		&updatedPost.Title,
		&updatedPost.Content,
		&updatedPost.CreatedAt,
		&updatedPost.UpdatedAt,
	)

	// Обрабатываем ошибки
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return updatedPost, nil
}

// Удаляем пост по ID, проверяем что запись существовала
func (r *PostgresPostRepository) DeletePost(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Проверяем, что пост был удален (RowsAffected > 0)
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// Возвращаем список постов с пагинацией (limit/offset)
func (r *PostgresPostRepository) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	query := `
        SELECT id, author_id, title, content, created_at, updated_at
        FROM posts 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		// Сканируем каждую строку результата
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content,
			&post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// Подсчитывает общее количество постов (для пагинации)
func (r *PostgresPostRepository) CountPosts(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return count, nil
}

// Список постов конкретного пользователя с пагинацией
func (r *PostgresPostRepository) ListPostsByUser(ctx context.Context, userID, limit, offset int) ([]*model.Post, error) {
	query := `
        SELECT id, author_id, title, content, created_at, updated_at
        FROM posts 
        WHERE author_id = $1
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content,
			&post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}
