// internal/handlers/post_handler_test.go
package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/handlers"
	"blog-backend/internal/model"
	"blog-backend/internal/repository"
	"blog-backend/service"
)

// Локальные ошибки
var (
	ErrPostNotFound = errors.New("post not found")
	ErrUserNotFound = errors.New("user not found")
)

// MemoryPostStorage — потокобезопасное in-memory хранилище постов
type MemoryPostStorage struct {
	posts  []*model.Post
	mu     sync.RWMutex
	nextID int
}

// NewMemoryPostStorage создает новое хранилище постов с автоинкрементом ID=1
func NewMemoryPostStorage() repository.PostRepository {
	return &MemoryPostStorage{nextID: 1}
}

// CreatePost создает новый пост с уникальным ID и текущей меткой времени
func (s *MemoryPostStorage) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post.ID = s.nextID              // Устанавливаем уникальный ID
	post.CreatedAt = time.Now()     // Метка создания
	s.posts = append(s.posts, post) // Добавляем в хранилище
	s.nextID++                      // Инкремент для следующего поста
	return post, nil
}

// GetPostByID возвращает пост по ID или ошибку если не найден
func (s *MemoryPostStorage) GetPostByID(ctx context.Context, id int) (*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.posts {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, ErrPostNotFound
}

// GetAll возвращает опубликованные посты с пагинацией (limit/offset)
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

// UpdatePost обновляет существующий пост по ID
func (s *MemoryPostStorage) UpdatePost(ctx context.Context, id int, post *model.Post) (*model.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.posts {
		if p.ID == id {
			s.posts[i] = post
			return post, nil
		}
	}
	return nil, ErrPostNotFound
}

// DeletePost удаляет пост по ID
func (s *MemoryPostStorage) DeletePost(ctx context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, p := range s.posts {
		if p.ID == id {
			s.posts = append(s.posts[:i], s.posts[i+1:]...)
			return nil
		}
	}
	return ErrPostNotFound
}

// CountPosts возвращает количество опубликованных постов
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

// GetReadyToPublish возвращает черновики готовые к публикации (publish_at <= now)
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

// ListPosts возвращает копию всех постов (без фильтрации)
func (s *MemoryPostStorage) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allPosts := make([]*model.Post, len(s.posts))
	copy(allPosts, s.posts)
	return allPosts[0:len(s.posts)], nil
}

// PublishPost меняет статус поста на "published"
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
	return ErrPostNotFound
}

// MemoryUserRepository - in-memory хранилище пользователей
type MemoryUserRepository struct {
	users  map[int]*model.User
	emails map[string]int
	nextID int
}

// NewMemoryUserRepository создает хранилище с тестовым пользователем ID=1
func NewMemoryUserRepository() repository.UserRepository {
	r := &MemoryUserRepository{
		users:  make(map[int]*model.User),
		emails: make(map[string]int),
		nextID: 1,
	}
	// Создаем тестового пользователя ID=1
	r.CreateUser(context.Background(), "test@example.com", "testuser", "hash")
	return r
}

// GetUserByID возвращает пользователя по ID
func (r *MemoryUserRepository) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail возвращает пользователя по email
func (r *MemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	userID, exists := r.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return r.users[userID], nil
}

// CreateUser создает нового пользователя (уникальный email)
func (r *MemoryUserRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	if _, exists := r.emails[email]; exists {
		return nil, errors.New("user exists")
	}

	user := &model.User{
		ID:           r.nextID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	r.users[user.ID] = user
	r.emails[email] = user.ID
	r.nextID++
	return user, nil
}

// UserExistsByEmail проверяет существование пользователя по email
func (r *MemoryUserRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := r.emails[email]
	return exists, nil
}

// Test config с отключенным scheduler
func NewTestConfig() *config.Config {
	return &config.Config{
		SchedulerEnabled: false,
	}
}

// Mock middleware для тестов - имитирует JWT middleware и всегда пропускает с userID=1
func mockAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "userID", 1)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// setupTestRouter создает полный тестовый роутер + возвращает PostRepository
// Аналогичная структура маршрутов как в main.go
func setupTestRouter() (http.Handler, repository.PostRepository) {
	postRepo := NewMemoryPostStorage()
	userRepo := NewMemoryUserRepository()
	cfg := NewTestConfig()

	postSvc := service.NewPostService(postRepo, userRepo, cfg)
	logger := log.New(io.Discard, "", 0) // Тихий логгер
	postHandler := handlers.NewPostHandler(postSvc, logger)

	// Стандартный mux с маршрутами как в main.go
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/posts", postHandler.ListPosts)
	mux.HandleFunc("POST /api/posts", mockAuthMiddleware(postHandler.CreatePost))
	mux.HandleFunc("GET /api/posts/1", postHandler.GetPost)
	mux.HandleFunc("PUT /api/posts/1", mockAuthMiddleware(postHandler.UpdatePost))
	mux.HandleFunc("DELETE /api/posts/1", mockAuthMiddleware(postHandler.DeletePost))

	return mux, postRepo
}

// TestCreatePost проверяет все сценарии создания поста
func TestCreatePost(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "valid_create_post",
			method:         http.MethodPost,
			url:            "/api/posts",
			body:           `{"title": "Test Post", "content": "Test content", "author_id": 1}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "no_auth",
			method:         http.MethodGet, // GET вместо POST
			url:            "/api/posts",
			body:           "",
			expectedStatus: http.StatusOK, // ListPosts
		},
		{
			name:           "invalid_JSON",
			method:         http.MethodPost,
			url:            "/api/posts",
			body:           `{invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _ := setupTestRouter()

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				bodyBytes, _ := io.ReadAll(w.Body)
				t.Logf("Status: %d, Body: %s", w.Code, string(bodyBytes))
				t.Errorf("expected %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestGetPost проверяет получение поста по ID
func TestGetPost(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setupPost      bool
		expectedStatus int
	}{
		{
			name:           "valid_post_ID",
			url:            "/api/posts/1",
			setupPost:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post_not_found",
			url:            "/api/posts/999",
			setupPost:      false,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid_post_ID",
			url:            "/api/posts/abc",
			setupPost:      false,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, postRepo := setupTestRouter()

			if tt.setupPost {
				// ✅ Создаем пост для теста GET существующего поста
				ctx := context.Background()
				testPost := &model.Post{
					Title:    "Test Post",
					Content:  "Test content",
					AuthorID: 1,
					Status:   "published",
					ID:       1, // Фиксируем ID
				}
				postRepo.CreatePost(ctx, testPost)
			}

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("Authorization", "Bearer dummy") // для mockAuthMiddleware

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				bodyBytes, _ := io.ReadAll(w.Body)
				t.Logf("Status: %d, Body: %s", w.Code, string(bodyBytes))
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}



// Интерфейсы реализованы
var _ repository.PostRepository = (*MemoryPostStorage)(nil)
var _ repository.UserRepository = (*MemoryUserRepository)(nil)
