package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectPostgres() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse postgres config error: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create postgres pool error: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("postgres ping error: %w", err)
	}

	DB = pool
	log.Println("[INFO] Connected to PostgreSQL")
	return nil
}

func CreateTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id            BIGSERIAL PRIMARY KEY,
		email         TEXT NOT NULL UNIQUE,
		username      TEXT NOT NULL UNIQUE,
		password      TEXT NOT NULL,
		created_at    TIMESTAMP NOT NULL DEFAULT now(),
		updated_at    TIMESTAMP NOT NULL DEFAULT now()
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';
	ALTER TABLE users ADD COLUMN IF NOT EXISTS banner_url TEXT NOT NULL DEFAULT '';

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id         BIGSERIAL PRIMARY KEY,
		user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT now()
	);

	CREATE TABLE IF NOT EXISTS posts (
		id         BIGSERIAL PRIMARY KEY,
		user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		username   TEXT NOT NULL,
		title      TEXT NOT NULL,
		content    TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT now()
	);

	CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);

	CREATE TABLE IF NOT EXISTS follows (
		follower_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		following_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at   TIMESTAMP NOT NULL DEFAULT now(),
		PRIMARY KEY (follower_id, following_id),
		CHECK (follower_id <> following_id)
	);

	CREATE INDEX IF NOT EXISTS idx_follows_following_id ON follows (following_id);`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := DB.Exec(ctx, schema); err != nil {
		return fmt.Errorf("create tables error: %w", err)
	}

	log.Println("[INFO] Tables created")
	return nil
}
