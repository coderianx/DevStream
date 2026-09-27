package users

import (
	"backend/internal/middlewares"
	"backend/internal/store"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const (
	maxAvatarSize = 2 << 20
	maxBannerSize = 4 << 20
)

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
	})
}

func saveUpload(w http.ResponseWriter, r *http.Request, field string, maxSize int64) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize+(1<<20))

	file, header, err := r.FormFile(field)
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			writeError(w, http.StatusRequestEntityTooLarge, "File is too large.")
			return "", false
		}
		writeError(w, http.StatusBadRequest, field+" file is required")
		return "", false
	}
	defer file.Close()

	if header.Size > maxSize {
		writeError(w, http.StatusRequestEntityTooLarge, "File is too large.")
		return "", false
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Println("[ERROR] upload read:", err)
		writeError(w, http.StatusBadRequest, "Invalid file")
		return "", false
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Println("[ERROR] upload seek:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return "", false
	}

	var ext string
	switch http.DetectContentType(buf[:n]) {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		writeError(w, http.StatusUnsupportedMediaType, "Only JPEG, PNG and WebP images are allowed")
		return "", false
	}

	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Println("[ERROR] upload dir:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return "", false
	}

	rnd := make([]byte, 16)
	if _, err := rand.Read(rnd); err != nil {
		log.Println("[ERROR] random name:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return "", false
	}
	name := hex.EncodeToString(rnd) + ext

	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		log.Println("[ERROR] upload create:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return "", false
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Println("[ERROR] upload save:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return "", false
	}

	return "/uploads/" + name, true
}

func UploadAvatar(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	avatarURL, ok := saveUpload(w, r, "avatar", maxAvatarSize)
	if !ok {
		return
	}

	if _, err := store.DB.Exec(r.Context(),
		`UPDATE users SET avatar_url = $1, updated_at = now() WHERE id = $2`,
		avatarURL, claims.UserID,
	); err != nil {
		log.Println("[ERROR] avatar update:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"avatar_url": avatarURL,
	})
}

func UploadBanner(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	bannerURL, ok := saveUpload(w, r, "banner", maxBannerSize)
	if !ok {
		return
	}

	if _, err := store.DB.Exec(r.Context(),
		`UPDATE users SET banner_url = $1, updated_at = now() WHERE id = $2`,
		bannerURL, claims.UserID,
	); err != nil {
		log.Println("[ERROR] banner update:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"banner_url": bannerURL,
	})
}

func resolveUserIDByUsername(ctx context.Context, username string) (int64, error) {
	var id int64
	err := store.DB.QueryRow(ctx,
		`SELECT id FROM users WHERE username = $1`,
		username,
	).Scan(&id)
	return id, err
}

func FollowUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetID, err := resolveUserIDByUsername(ctx, chi.URLParam(r, "username"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Println("[ERROR] follow target select:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if targetID == claims.UserID {
		writeError(w, http.StatusBadRequest, "You cannot follow yourself")
		return
	}

	if _, err := store.DB.Exec(ctx,
		`INSERT INTO follows (follower_id, following_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		claims.UserID, targetID,
	); err != nil {
		log.Println("[ERROR] follow insert:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"following": true,
	})
}

func UnfollowUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetID, err := resolveUserIDByUsername(ctx, chi.URLParam(r, "username"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Println("[ERROR] unfollow target select:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if targetID == claims.UserID {
		writeError(w, http.StatusBadRequest, "You cannot unfollow yourself")
		return
	}

	if _, err := store.DB.Exec(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		claims.UserID, targetID,
	); err != nil {
		log.Println("[ERROR] unfollow delete:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"following": false,
	})
}

func FollowState(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	claims, ok := middlewares.GetClaims(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetID, err := resolveUserIDByUsername(ctx, chi.URLParam(r, "username"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Println("[ERROR] follow state target select:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var following bool
	if err := store.DB.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM follows
			WHERE follower_id = $1 AND following_id = $2
		 )`,
		claims.UserID, targetID,
	).Scan(&following); err != nil {
		log.Println("[ERROR] follow state select:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"following": following,
	})
}

func SearchUsers(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		json.NewEncoder(w).Encode(map[string]any{
			"users": []any{},
		})
		return
	}

	pattern := strings.ReplaceAll(q, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")

	rows, err := store.DB.Query(r.Context(),
		`SELECT username, avatar_url
		 FROM users
		 WHERE username ILIKE $1 || '%'
		 ORDER BY username
		 LIMIT 10`,
		pattern,
	)
	if err != nil {
		log.Println("[ERROR] search users:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	users := make([]map[string]any, 0)
	for rows.Next() {
		var username, avatarURL string
		if err := rows.Scan(&username, &avatarURL); err != nil {
			log.Println("[ERROR] search scan:", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		users = append(users, map[string]any{
			"username":   username,
			"avatar_url": avatarURL,
		})
	}
	if err := rows.Err(); err != nil {
		log.Println("[ERROR] search rows:", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"users": users,
	})
}
