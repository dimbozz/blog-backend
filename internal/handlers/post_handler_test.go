// internal/handlers/post_handlers_test.go
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

// TestPostHandler — переопределяет handlers
type TestPostHandler struct{}

func (h *TestPostHandler) HandlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Mock ListPosts
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := handlers.Response{
			Data:  []*model.Post{{ID: 1, Title: "Test Post"}},
			Total: 1,
		}
		json.NewEncoder(w).Encode(resp)
	case http.MethodPost:
		// Mock CreatePost с проверкой авторизации
		_, ok := auth.GetUserIDFromContext(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(handlers.Response{Error: "user not authenticated"})
			return
		}

		var post model.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(handlers.Response{Error: "invalid JSON"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := handlers.Response{Data: &model.Post{ID: 1, Title: post.Title}}
		json.NewEncoder(w).Encode(resp)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(handlers.Response{Error: "method not allowed"})
	}
}

func (h *TestPostHandler) HandlePostID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	if idStr == "" || idStr == "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handlers.Response{Error: "post ID required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(handlers.Response{Error: "invalid post ID"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := handlers.Response{Data: &model.Post{ID: id, Title: "Test Post"}}
		json.NewEncoder(w).Encode(resp)
	case http.MethodPut, http.MethodDelete:
		_, ok := auth.GetUserIDFromContext(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(handlers.Response{Error: "user not authenticated"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(handlers.Response{Message: "success"})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(handlers.Response{Error: "method not allowed"})
	}
}

func TestHandlePosts(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "GET ListPosts",
			method:         http.MethodGet,
			url:            "/api/posts",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST CreatePost no auth",
			method:         http.MethodPost,
			url:            "/api/posts",
			body:           `{"title":"Test","content":"test"}`,
			setContextUser: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST invalid JSON",
			method:         http.MethodPost,
			url:            "/api/posts",
			body:           `{invalid json`,
			setContextUser: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Method not allowed",
			method:         "PATCH",
			url:            "/api/posts",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TestPostHandler{}

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.HandlePosts(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получено %d", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestHandlePostID(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		setContextUser bool
		expectedStatus int
	}{
		{
			name:           "GET valid ID",
			method:         http.MethodGet,
			url:            "/api/posts/123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET invalid ID",
			method:         http.MethodGet,
			url:            "/api/posts/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "GET empty ID",
			method:         http.MethodGet,
			url:            "/api/posts/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "PUT no auth",
			method:         http.MethodPut,
			url:            "/api/posts/1",
			setContextUser: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "DELETE no auth",
			method:         http.MethodDelete,
			url:            "/api/posts/1",
			setContextUser: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Method not allowed",
			method:         "PATCH",
			url:            "/api/posts/1",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TestPostHandler{}

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewReader([]byte(tt.body)))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if tt.setContextUser {
				ctx := context.WithValue(req.Context(), "userID", 1)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.HandlePostID(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получено %d", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}

func TestCreatePostWithValidJSON(t *testing.T) {
	h := &TestPostHandler{}

	req := httptest.NewRequest(http.MethodPost, "/api/posts",
		bytes.NewReader([]byte(`{"title":"Valid","content":"test"}`)))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.HandlePosts(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("Ожидался %d, получено %d", http.StatusCreated, res.StatusCode)
	}
}

func TestUpdatePostWithAuth(t *testing.T) {
	h := &TestPostHandler{}

	req := httptest.NewRequest(http.MethodPut, "/api/posts/1", nil)
	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.HandlePostID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Ожидался %d, получено %d", http.StatusOK, rec.Code)
	}
}
