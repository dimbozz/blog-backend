// service/comment_service.go
package service

import (
	"context"
	"fmt"

	"blog-backend/internal/model"
	"blog-backend/internal/repository"
)

type CommentService struct {
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	userRepo repository.UserRepository,
) *CommentService {
	return &CommentService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
	}
}

// CreateComment создает комментарий с проверками
func (s *CommentService) CreateComment(ctx context.Context, userID int, postID int, content string) (*model.Comment, error) {
	// 1. Проверяем существование поста
	if _, err := s.postRepo.GetPostByID(ctx, postID); err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// 2. Проверяем существование пользователя
	if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 3. Валидация контента
	if content == "" || len(content) > 1000 {
		return nil, fmt.Errorf("content required and max 1000 chars")
	}

	// 4. Создаем комментарий
	comment := &model.Comment{
		PostID:   postID,
		AuthorID: userID,
		Content:  content,
	}

	id, err := s.commentRepo.Create(ctx, comment)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// 5. Возвращаем созданный комментарий
	createdComment, err := s.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created comment: %w", err)
	}

	// Находим только что созданный
	for _, c := range createdComment {
		if c.ID == id {
			return c, nil
		}
	}

	return nil, fmt.Errorf("created comment not found")
}

// GetCommentsByPostID возвращает комментарии поста
func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID int) ([]*model.Comment, error) {
	// Проверяем существование поста
	if _, err := s.postRepo.GetPostByID(ctx, postID); err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// Получаем комментарии
	comments, err := s.commentRepo.GetByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	return comments, nil
}
