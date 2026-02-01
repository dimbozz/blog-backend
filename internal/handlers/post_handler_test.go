// internal/handlers/post_handlers_test.go
package handlers_test

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-backend/internal/handlers"
)

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
			expectedStatus: http.StatusInternalServerError,
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
			body:           `invalid json`,
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
			h := handlers.NewPostHandler(nil, log.New(io.Discard, "", 0))

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
			expectedStatus: http.StatusInternalServerError,
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
			body:           `{"title":"Updated"}`,
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
			h := handlers.NewPostHandler(nil, log.New(io.Discard, "", 0))

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

func TestGetPost(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "valid ID",
			url:            "/api/posts/123",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid ID",
			url:            "/api/posts/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty ID",
			url:            "/api/posts/",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewPostHandler(nil, log.New(io.Discard, "", 0))

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			h.GetPost(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получено %d", tt.expectedStatus, res.StatusCode)
			}
		})
	}
}
