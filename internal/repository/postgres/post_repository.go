package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"blog-backend/internal/model"
)

type PostgresPostRepository struct {
	db *sql.DB
}

func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	query := `
        INSERT INTO posts (user_id, title, content) 
        VALUES ($1, $2, $3) 
        RETURNING id, user_id, title, content, created_at, updated_at`

	// Инициализируем структуру Post
	createdPost := &model.Post{}

	// Выполняем запрос INSERT INTO
	row := r.db.QueryRowContext(ctx, query, post.UserID, post.Title, post.Content)

	// Заполняем поля структуры createdPost данными из БД
	err := row.Scan(
		&createdPost.ID,        // int
		&createdPost.UserID,    // int
		&createdPost.Title,     // string
		&createdPost.Content,   // string
		&createdPost.CreatedAt, // time.Time
		&createdPost.UpdatedAt, // time.Time
	)

	// Обрабатываем ошибки
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return createdPost, nil
}

func (r *PostgresPostRepository) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	post := &model.Post{}
	query := `
        SELECT id, user_id, title, content, created_at, updated_at 
        FROM posts 
        WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return post, nil
}

func (r *PostgresPostRepository) UpdatePost(ctx context.Context, id int, post *model.Post) (*model.Post, error) {
	query := `
        UPDATE posts 
        SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, user_id, title, content, created_at, updated_at`

	updatedPost := &model.Post{}

	row := r.db.QueryRowContext(ctx, query, post.Title, post.Content, id)

	err := row.Scan(
		&updatedPost.ID,
		&updatedPost.UserID,
		&updatedPost.Title,
		&updatedPost.Content,
		&updatedPost.CreatedAt,
		&updatedPost.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return updatedPost, nil
}

func (r *PostgresPostRepository) DeletePost(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *PostgresPostRepository) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	query := `
        SELECT id, user_id, title, content, created_at, updated_at
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
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content,
			&post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostgresPostRepository) CountPosts(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count posts: %w", err)
	}
	return count, nil
}

func (r *PostgresPostRepository) ListPostsByUser(ctx context.Context, userID, limit, offset int) ([]*model.Post, error) {
	query := `
        SELECT id, user_id, title, content, created_at, updated_at
        FROM posts 
        WHERE user_id = $1
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
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content,
			&post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	return posts, nil
}
