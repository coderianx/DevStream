package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func ConnectRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return fmt.Errorf("REDIS_ADDR environment variable is not set")
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}

	log.Println("[INFO] Connected to Redis")
	return nil
}
