// handlers/comment.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		http.Error(w, "invalid post ID CreateComment", http.StatusBadRequest)
		return
	}

	// userID из JWT middleware
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

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

// GET /api/posts/{postId}/comments
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// postId автоматически из пути
	postId := r.PathValue("postId")
	postID, err := strconv.Atoi(postId)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.commentSvc.GetCommentsByPostID(r.Context(), postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
