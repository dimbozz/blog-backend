// internal/repository/postgres/comment_repository.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"blog-backend/internal/model"
)

type CommentRepository struct {
	db *sql.DB
}

func NewPostgresCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create сохраняет комментарий и возвращает ID
func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) (int, error) {
	query := `
        INSERT INTO comments (post_id, author_id, content, created_at)
        VALUES ($1, $2, $3, NOW())
        RETURNING id`

	var id int
	err := r.db.QueryRowContext(ctx, query,
		comment.PostID,
		comment.AuthorID,
		comment.Content,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create comment: %w", err)
	}

	comment.ID = id
	return id, nil
}

// GetByPostID возвращает все комментарии поста
func (r *CommentRepository) GetByPostID(ctx context.Context, postID int) ([]*model.Comment, error) {
	query := `
        SELECT id, post_id, author_id, content, created_at 
        FROM comments 
        WHERE post_id = $1 
        ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post_id %d: %w", postID, err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		comment := &model.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Content,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return comments, nil
}
