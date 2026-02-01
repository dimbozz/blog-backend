// service_test/memory_post.go
// реализация in-memory хранилища постов для тестов
package service_test

import (
	"context"
	"errors"
	"sync"
	"time"

	"blog-backend/internal/model"
	"blog-backend/internal/repository" // интерфейс PostRepository
)

// MemoryPostStorage — потокобезопасное in-memory хранилище постов
type MemoryPostStorage struct {
	posts  []*model.Post // список всех постов
	mu     sync.RWMutex  // RWMutex для потокобезопасности
	nextID int           // автоинкрементный ID
}

// NewMemoryPostStorage создает новое хранилище и возвращает интерфейс PostRepository
func NewMemoryPostStorage() repository.PostRepository {
	return &MemoryPostStorage{nextID: 1}
}

// Create создает новый пост с уникальным ID
func (s *MemoryPostStorage) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post.ID = s.nextID
	post.CreatedAt = time.Now()
	s.posts = append(s.posts, post)
	s.nextID++
	return post, nil
}

// Get возвращает пост по ID (thread-safe)
func (s *MemoryPostStorage) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.posts {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New("post not found")
}

// GetAll возвращает опубликованные посты с пагинацией
func (s *MemoryPostStorage) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var published []*model.Post
	for _, post := range s.posts {
		if post.Status == "published" {
			published = append(published, post)
		}
	}

	total := len(published)
	if offset >= total {
		return nil, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return published[offset:end], total, nil
}

// Update обновляет существующий пост
func (s *MemoryPostStorage) UpdatePost(ctx context.Context, id int, post *model.Post) (*model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.posts {
		if p.ID == post.ID {
			s.posts[i] = post
			return post, nil
		}
	}
	return post, errors.New("post not found")
}

// Delete удаляет пост по ID
func (s *MemoryPostStorage) DeletePost(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.posts {
		if p.ID == id {
			s.posts = append(s.posts[:i], s.posts[i+1:]...)
			return nil
		}
	}
	return errors.New("post not found")
}

// Возвращает количество опубликованных постов
func (s *MemoryPostStorage) CountPosts(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, post := range s.posts {
		if post.Status == "published" {
			count++
		}
	}
	return count, nil
}

// GetReadyToPublish возвращает посты готовые к публикации (publish_at <= now)
func (s *MemoryPostStorage) GetReadyToPublish(ctx context.Context, batchSize int) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var ready []*model.Post
	for _, post := range s.posts {
		if post.Status == "draft" && !post.PublishAt.IsZero() && post.PublishAt.Before(now) {
			ready = append(ready, post)
		}
	}
	return ready, nil
}

// Метод ListPosts
func (s *MemoryPostStorage) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Возвращаем КОПИЮ всех постов (потоко-безопасно)
	allPosts := make([]*model.Post, len(s.posts))
	copy(allPosts, s.posts)
	return allPosts, nil
}

// PublishPost — публикует пост по ID
func (s *MemoryPostStorage) PublishPost(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post.ID == id {
			post.Status = "published"
			s.posts[i] = post
			return nil
		}
	}
	return errors.New("post not found")
}

// Проверка — все методы реализованы
var _ repository.PostRepository = (*MemoryPostStorage)(nil)
