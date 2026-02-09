package handlers

import (
	"blog-backend/internal/handlers/middleware"
	"blog-backend/internal/model"
	"blog-backend/pkg/auth"
	"blog-backend/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Response - единый JSON формат ответа API
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Total   int         `json:"total,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PostHandler - HTTP обработчики для постов
type PostHandler struct {
	postService *service.PostService
	log         *log.Logger
}

// NewPostHandler создает новый PostHandler
func NewPostHandler(postService *service.PostService, logger *log.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		log:         logger,
	}
}

// CreatePost создает новый пост (требуется авторизация)
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r)
	if !ok {
		middleware.AbortError(w, r, "User not authenticated", http.StatusUnauthorized, nil)
		return
	}

	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		middleware.AbortError(w, r, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	createdPost, err := h.postService.CreatePost(r.Context(), userID, &post)
	if err != nil {
		middleware.AbortError(w, r, "Failed to create post", http.StatusInternalServerError, err)
		return
	}

	h.successResponse(w, http.StatusCreated, Response{
		Data: createdPost,
	})
}

// GetPost возвращает пост по ID (публичный доступ)
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	if idStr == "" || idStr == "/" {
		middleware.AbortError(w, r, "Post ID required", http.StatusBadRequest, nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.AbortError(w, r, "Invalid post ID", http.StatusBadRequest, err)
		return
	}

	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		middleware.AbortError(w, r, "Post not found", http.StatusNotFound, err)
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data: post,
	})
}

// UpdatePost обновляет пост (только автор)
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r)
	if !ok {
		middleware.AbortError(w, r, "User not authenticated", http.StatusUnauthorized, nil)
		return
	}

	// Парсим ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	if idStr == "" || idStr == "/" {
		middleware.AbortError(w, r, "Post ID required", http.StatusBadRequest, nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.AbortError(w, r, "Invalid post ID", http.StatusBadRequest, err)
		return
	}

	// Парсим поля для обновления (кроме ID)
	var updateData struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		middleware.AbortError(w, r, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	// Создаем post с ID из URL
	postToUpdate := &model.Post{
		ID:      id,
		Title:   updateData.Title,
		Content: updateData.Content,
		Status:  updateData.Status,
	}

	updatedPost, err := h.postService.UpdatePost(r.Context(), userID, id, postToUpdate)
	if err != nil {
		// Ловим ВСЕ ошибки сервиса
		switch {
		case strings.Contains(err.Error(), "post not found"):
			middleware.AbortError(w, r, "Post not found", http.StatusNotFound, err)
		case strings.Contains(err.Error(), "permission denied"):
			middleware.AbortError(w, r, "Permission denied", http.StatusForbidden, err)
		default:
			middleware.AbortError(w, r, "Failed to update post", http.StatusInternalServerError, err)
		}
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data:    updatedPost,
		Message: "post updated successfully",
	})
}

// DeletePost удаляет пост (только автор)
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {

	userID, ok := auth.GetUserIDFromContext(r)
	if !ok {
		middleware.AbortError(w, r, "User not authenticated", http.StatusUnauthorized, nil)
		return
	}

	// Парсим ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	idStr = strings.TrimSuffix(idStr, "/")
	if idStr == "" || idStr == "/" {
		middleware.AbortError(w, r, "Post ID required", http.StatusBadRequest, nil)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.AbortError(w, r, "Invalid post ID", http.StatusBadRequest, err)
		return
	}

	// Обработка ошибок сервиса
	if err := h.postService.DeletePost(r.Context(), userID, id); err != nil {
		switch {
		case strings.Contains(err.Error(), "post not found"):
			middleware.AbortError(w, r, "Post not found", http.StatusNotFound, err)
		case strings.Contains(err.Error(), "permission denied"):
			middleware.AbortError(w, r, "Permission denied", http.StatusForbidden, err)
		default:
			middleware.AbortError(w, r, "Failed to delete post", http.StatusInternalServerError, err)
		}
		return
	}

	// 204 No Content для DELETE
	w.WriteHeader(http.StatusNoContent)
}

// ListPosts возвращает все посты с пагинацией
func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset == 0 {
		offset = 0
	}

	posts, total, err := h.postService.GetAllPosts(r.Context(), limit, offset)
	if err != nil {
		h.log.Printf("list posts failed: %v", err)
		middleware.AbortError(w, r, "Failed to list posts", http.StatusInternalServerError, err)
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data:  posts,
		Total: total,
	})
}

// successResponse отправляет успешный JSON ответ
func (h *PostHandler) successResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
