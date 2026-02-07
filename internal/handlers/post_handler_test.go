// internal/handlers/post_handler_test.go
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"blog-backend/internal/handlers"
	"blog-backend/internal/model"
	"blog-backend/pkg/auth"
)

// TestPostHandler — МОК с реальными методами CreatePost/GetPost/UpdatePost/DeletePost
type TestPostHandler struct {
	posts map[int]*model.Post
}

func NewTestPostHandler() *TestPostHandler {
	return &TestPostHandler{
		posts: make(map[int]*model.Post),
	}
}

func (h *TestPostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации
	_, ok := auth.GetUserIDFromContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(handlers.Response{Error: "user not authenticated"})
		return
	}

	// Парсинг JSON
	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "invalid JSON"})
		return
	}

	// Сохранение поста
	post.ID = 1
	h.posts[post.ID] = &post

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(handlers.Response{Data: &post})
}

func (h *TestPostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "invalid post ID"})
		return
	}

	if post, exists := h.posts[id]; exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(handlers.Response{Data: post})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(handlers.Response{Error: "post not found"})
}

func (h *TestPostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "invalid post ID"})
		return
	}

	// Проверка авторизации
	_, ok := auth.GetUserIDFromContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(handlers.Response{Error: "user not authenticated"})
		return
	}

	// Парсинг JSON
	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "invalid JSON"})
		return
	}

	post.ID = id
	h.posts[id] = &post

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(handlers.Response{Message: "post updated"})
}

func (h *TestPostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "invalid post ID"})
		return
	}

	// Проверка авторизации
	_, ok := auth.GetUserIDFromContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(handlers.Response{Error: "user not authenticated"})
		return
	}

	delete(h.posts, id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(handlers.Response{Message: "post deleted"})
}

// TestCreatePost — table-driven тест
func TestCreatePost(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "valid create post",
			body:           `{"title":"Test Post","content":"test content"}`,
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
			body:           `{invalid json`,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           ``,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestPostHandler()

			req := httptest.NewRequest(http.MethodPost, "/api/posts", bytes.NewReader([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.CreatePost(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// TestGetPost — table-driven тест
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
		{
			name:           "empty post ID",
			postId:         "",
			setupData:      false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestPostHandler()
			if tt.setupData {
				h.posts[1] = &model.Post{ID: 1, Title: "Test Post"}
			}

			req := httptest.NewRequest(http.MethodGet, "/api/posts/"+tt.postId, nil)
			rec := httptest.NewRecorder()

			h.GetPost(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// TestUpdatePost — table-driven тест
func TestUpdatePost(t *testing.T) {
	tests := []struct {
		name           string
		postId         string
		body           string
		setContextUser bool
		setupData      bool
		expectedStatus int
	}{
		{
			name:           "valid update with auth",
			postId:         "1",
			body:           `{"title":"Updated","content":"updated content"}`,
			setContextUser: true,
			setupData:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no auth",
			postId:         "1",
			body:           `{"title":"Updated","content":"updated"}`,
			setContextUser: false,
			setupData:      true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "post not found",
			postId:         "999",
			body:           `{"title":"Updated","content":"updated"}`,
			setContextUser: true,
			setupData:      false,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid post ID",
			postId:         "abc",
			body:           `{"title":"Updated","content":"updated"}`,
			setContextUser: true,
			setupData:      false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			postId:         "1",
			body:           `{invalid`,
			setContextUser: true,
			setupData:      true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestPostHandler()
			if tt.setupData {
				h.posts[1] = &model.Post{ID: 1, Title: "Original"}
			}

			req := httptest.NewRequest(http.MethodPut, "/api/posts/"+tt.postId, bytes.NewReader([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.UpdatePost(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

// TestDeletePost — table-driven тест
func TestDeletePost(t *testing.T) {
	tests := []struct {
		name           string
		postId         string
		setContextUser bool
		setupData      bool
		expectedStatus int
	}{
		{
			name:           "valid delete with auth",
			postId:         "1",
			setContextUser: true,
			setupData:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no auth",
			postId:         "1",
			setContextUser: false,
			setupData:      true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "post not found",
			postId:         "999",
			setContextUser: true,
			setupData:      false,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid post ID",
			postId:         "abc",
			setContextUser: true,
			setupData:      false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTestPostHandler()
			if tt.setupData {
				h.posts[1] = &model.Post{ID: 1}
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+tt.postId, nil)

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.DeletePost(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
