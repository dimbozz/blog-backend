// internal/handlers/post_handler_test.go
package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

// MemoryUserRepository - ПОЛНАЯ реализация repository.UserRepository
type MemoryUserRepository struct {
	users  map[int]*model.User
	emails map[string]int // email -> userID
	nextID int
}

func NewMemoryUserRepository() repository.UserRepository {
	return &MemoryUserRepository{
		users:  make(map[int]*model.User),
		emails: make(map[string]int),
		nextID: 1,
	}
}

// ✅ GetUserByID - получение пользователя по ID
func (r *MemoryUserRepository) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ✅ GetUserByEmail - получение пользователя по email
func (r *MemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	userID, exists := r.emails[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return r.users[userID], nil
}

// ✅ CreateUser - создание пользователя
func (r *MemoryUserRepository) CreateUser(ctx context.Context, email, username, passwordHash string) (*model.User, error) {
	// Проверяем, существует ли уже email
	if _, exists := r.emails[email]; exists {
		return nil, errors.New("user with this email already exists")
	}

	user := &model.User{
		ID:           r.nextID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	r.nextID++
	r.users[user.ID] = user
	r.emails[email] = user.ID

	return user, nil
}

// ✅ UserExistsByEmail - проверка существования пользователя по email
func (r *MemoryUserRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := r.emails[email]
	return exists, nil
}

// createTestPostService создает сервис с memory хранилищем
func createTestPostService(t *testing.T) *service.PostService {
	postRepo := NewMemoryPostStorage()
	userRepo := NewMemoryUserRepository()
	cfg := NewTestConfig()
	_, cancel := context.WithCancel(context.Background())

	postSvc := service.NewPostService(postRepo, userRepo, cfg) // используем конструктор из пакета service

	t.Cleanup(cancel)
	return postSvc
}

// Минимальная тестовая конфигурация
func NewTestConfig() *config.Config {
	// Создаем пустую конфигурацию
	var cfg config.Config
	return &cfg
}

// createTestHandler создает полный mux с маршрутизацией
func createTestHandler(t *testing.T) http.Handler {
	postSvc := createTestPostService(t)

	var logBuffer bytes.Buffer
	logger := log.New(&logBuffer, "test: ", log.LstdFlags)

	postHandler := handlers.NewPostHandler(postSvc, logger)

	// Создаем mux для маршрутизации
	mux := http.NewServeMux()
	mux.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postHandler.CreatePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/posts" {
			return // обработано выше
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			ctx := context.WithValue(r.Context(), "postID", id)
			postHandler.GetPost(w, r.WithContext(ctx))
		case http.MethodPut:
			ctx := context.WithValue(r.Context(), "postID", id)
			postHandler.UpdatePost(w, r.WithContext(ctx))
		case http.MethodDelete:
			ctx := context.WithValue(r.Context(), "postID", id)
			postHandler.DeletePost(w, r.WithContext(ctx))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "valid create post",
			body:           `{"title":"Test","content":"test content"}`,
			setContextUser: true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "no auth",
			body:           `{"title":"Test","content":"test"}`,
			setContextUser: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid JSON",
			body:           `{invalid`,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createTestHandler(t)

			req := httptest.NewRequest(http.MethodPost, "/api/posts",
				bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		name           string
		postId         string
		setupData      bool
		expectedStatus int
	}{
		{
			name:           "valid post ID",
			postId:         "1",
			setupData:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "post not found",
			postId:         "999",
			setupData:      false,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid post ID",
			postId:         "abc",
			setupData:      false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := createTestHandler(t)

			if tt.setupData {
				// Создаем пост через POST запрос
				createReq := httptest.NewRequest(http.MethodPost, "/api/posts",
					bytes.NewReader([]byte(`{"title":"Test","content":"test"}`)))
				createReq.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(createReq.Context(), "userID", 1)
				createReq = createReq.WithContext(ctx)

				createRec := httptest.NewRecorder()
				handler.ServeHTTP(createRec, createReq)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/posts/"+tt.postId, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// Проверка — все методы реализованы
var _ repository.PostRepository = (*MemoryPostStorage)(nil)
var _ repository.UserRepository = (*MemoryUserRepository)(nil)
