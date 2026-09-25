package auth

import (
	"backend/internal/middlewars"
	"backend/internal/models"
	"backend/internal/store"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid request",
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Username == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Username, email and password are required",
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[ERROR] bcrypt hash:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	var id int64
	err = store.DB.QueryRow(ctx,
		`INSERT INTO users (email, username, password)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		req.Email, req.Username, string(hash),
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Username or email already exists",
			})
			return
		}
		log.Println("[ERROR] register insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":       id,
		"username": req.Username,
		"email":    req.Email,
	})
}

func Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid request",
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Username and password are required",
		})
		return
	}

	var (
		id       int64
		username string
		password string
	)
	err := store.DB.QueryRow(ctx,
		`SELECT id, username, password
		 FROM users
		 WHERE username = $1`,
		req.Username,
	).Scan(&id, &username, &password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Invalid credentials",
			})
			return
		}
		log.Println("[ERROR] login select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid credentials",
		})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[ERROR] JWT_SECRET is not set")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	claims := models.Claims{
		UserID:   id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println("[ERROR] token sign:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		log.Println("[ERROR] refresh token generate:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	_, err = store.DB.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		id, HashRefreshToken(refreshToken), time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		log.Println("[ERROR] refresh token insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"token":         tokenString,
		"refresh_token": refreshToken,
	})
}

func Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Invalid request",
		})
		return
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Refresh token is required",
		})
		return
	}

	var (
		userID   int64
		username string
	)
	err := store.DB.QueryRow(ctx,
		`SELECT rt.user_id, u.username
		 FROM refresh_tokens rt
		 JOIN users u ON u.id = rt.user_id
		 WHERE rt.token_hash = $1
		   AND rt.expires_at > now()`,
		HashRefreshToken(req.RefreshToken),
	).Scan(&userID, &username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Invalid or expired refresh token",
			})
			return
		}
		log.Println("[ERROR] refresh select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		log.Println("[ERROR] refresh token generate:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	newHash := HashRefreshToken(newRefreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tx, err := store.DB.Begin(ctx)
	if err != nil {
		log.Println("[ERROR] refresh tx begin:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, newHash, expiresAt,
	); err != nil {
		tx.Rollback(ctx)
		log.Println("[ERROR] refresh token insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		HashRefreshToken(req.RefreshToken),
	); err != nil {
		tx.Rollback(ctx)
		log.Println("[ERROR] refresh token delete:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("[ERROR] refresh tx commit:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("[ERROR] JWT_SECRET is not set")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	claims := models.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println("[ERROR] token sign:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"token":         tokenString,
		"refresh_token": newRefreshToken,
	})
}

func Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, ok := middlewars.GetClaims(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Unauthorized",
		})
		return
	}

	var (
		username  string
		email     string
		createdAt time.Time
	)
	err := store.DB.QueryRow(r.Context(),
		`SELECT username, email, created_at
		 FROM users
		 WHERE id = $1`,
		claims.UserID,
	).Scan(&username, &email, &createdAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "User not found",
			})
			return
		}
		log.Println("[ERROR] me select:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": "Internal server error",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":         claims.UserID,
		"username":   username,
		"email":      email,
		"created_at": createdAt,
	})
}
