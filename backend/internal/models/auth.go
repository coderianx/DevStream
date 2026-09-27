package models

import "github.com/golang-jwt/jwt/v5"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`

	jwt.RegisteredClaims
}
