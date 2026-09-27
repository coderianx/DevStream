package post

import (
	"backend/internal/middlewares"
	"backend/internal/models"
	"backend/internal/store"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	claims, ok := middlewares.GetClaims(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Unauthorized",
		})
		return
	}

	var req models.CreatePostRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid request",
		})
		return
	}

	if len(req.Title) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Title is required",
		})
		return
	}

	if len(req.Content) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Content is required",
		})
		return
	}

	if len(req.Title) > 255 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Title is too long",
		})
		return
	}

	if len(req.Content) > 10000 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Content is too long",
		})
		return
	}

	_, err = store.DB.Exec(
		ctx,
		`
		INSERT INTO posts (user_id, username, title, content)
		VALUES ($1, $2, $3, $4)
		`,
		claims.UserID,
		claims.Username,
		req.Title,
		req.Content,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Failed to create post",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Post created successfully",
	})
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Unauthorized",
		})
		return
	}

	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil || postID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid post ID",
		})
		return
	}

	tag, err := store.DB.Exec(ctx,
		`DELETE FROM posts WHERE id = $1 AND user_id = $2`,
		postID, claims.UserID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Failed to delete post",
		})
		return
	}

	if tag.RowsAffected() == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Post not found",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "Post deleted successfully",
	})
}
