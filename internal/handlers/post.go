package handlers

import (
	"blog-backend/internal/model"
	"blog-backend/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
func NewPostHandler(postService *service.PostService, log *log.Logger) *PostHandler {
	return &PostHandler{
		postService: postService,
		log:         log,
	}
}

// getUserIDFromContext извлекает userID из контекста (JWT middleware)
func (h *PostHandler) getUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}

// CreatePost создает новый пост (требуется авторизация)
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	createdPost, err := h.postService.CreatePost(r.Context(), userID, &post)
	if err != nil {
		h.log.Printf("create post failed: %v", err)
		h.errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.successResponse(w, http.StatusCreated, Response{
		Data: createdPost,
	})
}

// GetPost возвращает пост по ID (публичный доступ)
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "post ID required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	post, err := h.postService.GetPost(r.Context(), id)
	if err != nil {
		h.log.Printf("post not found: %d, err: %v", id, err)
		h.errorResponse(w, http.StatusNotFound, "post not found")
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data: post,
	})
}

// UpdatePost обновляет пост (только автор)
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "post ID required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	updatedPost, err := h.postService.UpdatePost(r.Context(), userID, id, &post)
	if err != nil {
		if err.Error() == "permission denied: can only update own posts" {
			h.errorResponse(w, http.StatusForbidden, "permission denied")
		} else {
			h.log.Printf("update post %d failed: %v", id, err)
			h.errorResponse(w, http.StatusInternalServerError, "internal server error")
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
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		h.errorResponse(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "post ID required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid post ID")
		return
	}

	if err := h.postService.DeletePost(r.Context(), userID, id); err != nil {
		if err.Error() == "permission denied: can only delete own posts" {
			h.errorResponse(w, http.StatusForbidden, "permission denied")
		} else {
			h.log.Printf("delete post %d failed: %v", id, err)
			h.errorResponse(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Message: "post deleted successfully",
	})
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
		h.errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data:  posts,
		Total: total,
	})
}

// ListUserPosts возвращает посты пользователя (публичный)
func (h *PostHandler) ListUserPosts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "user_id required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := h.postService.GetUserPosts(r.Context(), userID, limit, offset)
	if err != nil {
		h.log.Printf("list user %d posts failed: %v", userID, err)
		h.errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.successResponse(w, http.StatusOK, Response{
		Data: posts,
	})
}

// successResponse отправляет успешный JSON ответ
func (h *PostHandler) successResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// errorResponse отправляет ошибку в JSON формате
func (h *PostHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Error: message,
	})
}
