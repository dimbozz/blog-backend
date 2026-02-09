// handlers/comment.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"blog-backend/internal/handlers/middleware"
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

// POST /api/posts/{postId}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// postId автоматически из пути
	postId := r.PathValue("postId")
	postID, err := strconv.Atoi(postId)
	if err != nil {
		middleware.AbortError(w, r, "Invalid post ID", http.StatusBadRequest, err)
		return
	}

	// userID из JWT middleware
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		middleware.AbortError(w, r, "User not authenticated", http.StatusUnauthorized, nil)
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.AbortError(w, r, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	if req.Content == "" || len(req.Content) > 1000 {
		middleware.AbortError(w, r, "Content required, max 1000 chars", http.StatusBadRequest, nil)
		return
	}

	comment, err := h.commentSvc.CreateComment(r.Context(), userID, postID, req.Content)
	if err != nil {
		middleware.AbortError(w, r, "Failed to create comment", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateCommentResponse{
		ID:      comment.ID,
		Content: comment.Content,
	})
}

// GET /api/posts/{postId}/comments
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// postId автоматически из пути
	postId := r.PathValue("postId")
	postID, err := strconv.Atoi(postId)
	if err != nil {
		middleware.AbortError(w, r, "Invalid post ID", http.StatusBadRequest, err)
		return
	}

	comments, err := h.commentSvc.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		middleware.AbortError(w, r, "Failed to get comments", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
