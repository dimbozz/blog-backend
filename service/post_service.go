package service

import (
	"context"
	"fmt"

	"blog-backend/internal/model"
	"blog-backend/internal/repository"
)

// PostService - бизнес-логика постов (проверка прав + делегирование)
type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository // Для проверки пользователя
}

// Создаем сервис с репозиториями
func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// Создаем пост (текущий user = автор)
func (s *PostService) CreatePost(ctx context.Context, currentUserID int, post *model.Post) (*model.Post, error) {
	// Проверяем, что пользователь существует
	if _, err := s.userRepo.GetUserByID(ctx, currentUserID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Устанавливаем автора поста
	post.AuthorID = currentUserID

	// Делегируем в Repository
	return s.postRepo.CreatePost(ctx, post)
}

// Получаем пост по ID (для всех)
func (s *PostService) GetPost(ctx context.Context, id int) (*model.Post, error) {
	return s.postRepo.GetPostByID(ctx, id)
}

// Обновляет пост (только автор!)
func (s *PostService) UpdatePost(ctx context.Context, currentUserID, postID int, post *model.Post) (*model.Post, error) {
	// Получаем пост для проверки владельца
	existingPost, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// ПРОВЕРКА ПРАВ
	if existingPost.AuthorID != currentUserID {
		return nil, fmt.Errorf("permission denied: can only update own posts")
	}

	// Repository возвращает ОБНОВЛЕННЫЙ пост с updated_at из БД!
	updatedPost, err := s.postRepo.UpdatePost(ctx, postID, existingPost)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return updatedPost, nil
}

// ✅ Удаляет пост (только автор!)
func (s *PostService) DeletePost(ctx context.Context, currentUserID, postID int) error {
	// Проверяем права доступа
	existingPost, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	if existingPost.AuthorID != currentUserID {
		return fmt.Errorf("permission denied: can only delete own posts")
	}

	// Делегируем удаление
	return s.postRepo.DeletePost(ctx, postID)
}

// ✅ Все посты с пагинацией + total
func (s *PostService) GetAllPosts(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {
	// Получаем посты
	posts, err := s.postRepo.ListPosts(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts: %w", err)
	}

	// Получаем количество для пагинации
	total, err := s.postRepo.CountPosts(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	return posts, total, nil
}

// ✅ Посты конкретного пользователя
func (s *PostService) GetUserPosts(ctx context.Context, userID, limit, offset int) ([]*model.Post, error) {
	// Публичный доступ ко всем постам пользователя
	return s.postRepo.ListPostsByUser(ctx, userID, limit, offset)
}
