package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/model"
	"blog-backend/internal/repository"
)

// PostService - бизнес-логика постов (проверка прав + делегирование)
type PostService struct {
	postRepo     repository.PostRepository
	userRepo     repository.UserRepository // Для проверки пользователя
	wg           sync.WaitGroup
	ticker       *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
	workersCount int // Из .env
	batchSize    int // Из .env
}

// Создаем сервис с репозиториями
func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, cfg *config.Config) *PostService {
	ctx, cancel := context.WithCancel(context.Background())
	s := &PostService{
		postRepo:     postRepo,
		userRepo:     userRepo,
		ctx:          ctx,
		cancel:       cancel,
		ticker:       time.NewTicker(cfg.PostTickerDuration), // Из .env
		workersCount: cfg.PostWorkersCount,                   // Из .env
		batchSize:    cfg.PostBatchSize,                      // Из .env
	}

	s.StartScheduler()
	return s
}

// Запуск планировщика
func (s *PostService) StartScheduler() {
	s.wg.Add(1)
	go s.scheduler()
}

// Главная горутина планировщика (каждые N секунд из .env)
func (s *PostService) scheduler() {
	defer s.wg.Done()
	defer s.ticker.Stop()

	log.Printf("📅 Post scheduler started (every %v)", s.ticker.C)

	for {
		select {
		case <-s.ticker.C:
			s.publishPendingPosts()
		case <-s.ctx.Done():
			log.Println("📅 Post scheduler stopped")
			return
		}
	}
}

// Worker pool для конкурентной публикации
func (s *PostService) publishPendingPosts() {
	// 1. Берем готовые к публикации посты (max batchSize из .env)
	posts, err := s.postRepo.GetReadyToPublish(s.ctx, s.batchSize)
	if err != nil {
		log.Printf("Failed to get ready posts: %v", err)
		return
	}
	if len(posts) == 0 {
		return
	}

	log.Printf("Found %d posts ready to publish (max %d)", len(posts), s.batchSize)

	// 2. Канал для worker pool
	postChan := make(chan *model.Post, len(posts))
	for _, post := range posts {
		postChan <- post
	}
	close(postChan)

	// 3. Запускаем workersCount воркеров (из .env)
	var wg sync.WaitGroup
	for i := 0; i < s.workersCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.worker(postChan, workerID)
		}(i + 1)
	}
	wg.Wait()
}

// Воркер публикует один пост
func (s *PostService) worker(postChan <-chan *model.Post, workerID int) {
	for post := range postChan {
		if err := s.postRepo.PublishPost(s.ctx, post.ID); err != nil {
			log.Printf("Worker %d: failed to publish post %d: %v", workerID, post.ID, err)
		} else {
			log.Printf("Worker %d: published post %d (\"%s\")", workerID, post.ID, post.Title)
		}
	}
}

// Graceful shutdown
func (s *PostService) Stop() {
	log.Println("Stopping post service...")
	s.cancel()
	s.wg.Wait()
	log.Println("Post service stopped")
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
	if s.postRepo == nil {
		return nil, fmt.Errorf("postRepo is nil")
	}
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
	updatedPost, err := s.postRepo.UpdatePost(ctx, postID, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return updatedPost, nil
}

// Удаляет пост (только автор!)
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

// Все посты с пагинацией + total
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
