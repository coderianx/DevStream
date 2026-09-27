package public

import (
	"backend/internal/store"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func GetProfileByUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := chi.URLParam(r, "username")

	ctx := r.Context()

	var (
		id             int64
		avatarURL      string
		bannerURL      string
		createdAt      time.Time
		postsCount     int64
		followersCount int64
		followingCount int64
	)
	err := store.DB.QueryRow(ctx,
		`SELECT u.id, u.avatar_url, u.banner_url, u.created_at,
			(SELECT COUNT(*) FROM posts p WHERE p.user_id = u.id),
			(SELECT COUNT(*) FROM follows f WHERE f.following_id = u.id),
			(SELECT COUNT(*) FROM follows f WHERE f.follower_id = u.id)
		 FROM users u
		 WHERE u.username = $1`,
		username,
	).Scan(&id, &avatarURL, &bannerURL, &createdAt, &postsCount, &followersCount, &followingCount)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "User not found",
			})
			return
		}
		log.Println("[ERROR] public profile select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":              id,
		"username":        username,
		"avatar_url":      avatarURL,
		"banner_url":      bannerURL,
		"posts_count":     postsCount,
		"followers_count": followersCount,
		"following_count": followingCount,
		"created_at":      createdAt,
	})
}

func GetPostsByUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	username := chi.URLParam(r, "username")

	ctx := r.Context()

	var userID int64
	err := store.DB.QueryRow(ctx,
		`SELECT id FROM users WHERE username = $1`,
		username,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "User not found",
			})
			return
		}
		log.Println("[ERROR] posts user select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	rows, err := store.DB.Query(ctx,
		`SELECT id, title, content, created_at
		 FROM posts
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 50`,
		userID,
	)
	if err != nil {
		log.Println("[ERROR] posts select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}
	defer rows.Close()

	posts := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id        int64
			title     string
			content   string
			createdAt time.Time
		)
		if err := rows.Scan(&id, &title, &content, &createdAt); err != nil {
			log.Println("[ERROR] posts scan:", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Internal server error",
			})
			return
		}
		posts = append(posts, map[string]any{
			"id":         id,
			"title":      title,
			"content":    content,
			"created_at": createdAt,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERROR] posts rows:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"posts": posts,
	})
}
