package service

import (
	"blog-backend/internal/model"
	"blog-backend/internal/repository"
	"context"
)

type PostService struct {
	postRepo repository.PostRepository // интерфейс, не конкретная реализация
	userRepo repository.UserRepository // интерфейс для гибкости
}

// DTO
type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=1,max=200"`
	Content string `json:"content" validate:"required,min=10"`
}

func NewPostService(pr repository.PostRepository, ur repository.UserRepository) *PostService {
	return &PostService{
		postRepo: pr,
		userRepo: ur,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID int, req CreatePostRequest) (*model.Post, error) {
	// Проверяем, существует ли пользователь
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Создаём пост
	post := &model.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: user.ID,
	}

	// Сохраняем в БД
	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) GetPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	// Идём в БД
	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return posts, nil
}
