// handlers/comment.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"blog-backend/service"
)

type CommentHandler struct {
	commentSvc *service.CommentService
}

func NewCommentHandler(commentSvc *service.CommentService) *CommentHandler {
	return &CommentHandler{commentSvc: commentSvc}
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type CreateCommentResponse struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

// ServeHTTP обрабатывает все роуты комментариев
func (h *CommentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Парсим path: /api/posts/{postID}/comments/*
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api"), "/")
	if len(pathParts) < 4 || pathParts[0] != "posts" || pathParts[2] != "comments" {
		http.NotFound(w, r)
		return
	}

	postIDStr := pathParts[1]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	// userID из контекста (JWT middleware)
	userID, ok := r.Context().Value("userID").(int)
	if !ok && r.Method == "POST" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.handleCreateComment(w, r, postID, userID)
	case http.MethodGet:
		h.handleGetComments(w, r, postID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CommentHandler) handleCreateComment(w http.ResponseWriter, r *http.Request, postID int, userID int) {
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Content == "" || len(req.Content) > 1000 {
		http.Error(w, "content required, max 1000 chars", http.StatusBadRequest)
		return
	}

	comment, err := h.commentSvc.CreateComment(r.Context(), userID, postID, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateCommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
	})
}

func (h *CommentHandler) handleGetComments(w http.ResponseWriter, r *http.Request, postID int) {
	comments, err := h.commentSvc.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
