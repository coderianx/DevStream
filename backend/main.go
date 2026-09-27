package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"backend/internal/handlers/auth"
	"backend/internal/handlers/post"
	"backend/internal/handlers/public"
	"backend/internal/handlers/users"
	"backend/internal/middlewares"
	"backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERROR] .env file load error: ", err)
	}

	if err := store.ConnectPostgres(); err != nil {
		log.Fatal("[ERROR] postgres connect error: ", err)
	}
	defer store.DB.Close()

	if err := store.ConnectRedis(); err != nil {
		log.Fatal("[ERROR] redis connect error: ", err)
	}
	defer store.Redis.Close()

	if err := store.CreateTables(); err != nil {
		log.Fatal("[ERROR] create tables error: ", err)
	}

	router := chi.NewRouter()

	uploadsDir := os.Getenv("UPLOAD_DIR")
	if uploadsDir == "" {
		uploadsDir = "uploads"
	}
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		log.Fatal("[ERROR] create uploads dir error: ", err)
	}
	router.Handle("/uploads/*",
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))),
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.With(middlewares.RateLimitByIP(3, time.Minute)).Post("/auth/register", auth.Register)
		r.With(middlewares.RateLimitByIP(5, time.Minute)).Post("/auth/login", auth.Login)
		r.With(middlewares.RateLimitByIP(10, time.Minute)).Post("/auth/refresh", auth.Refresh)
		r.With(middlewares.RateLimitByIP(10, time.Minute)).Post("/auth/logout", auth.Logout)

		r.With(middlewares.RateLimitByIP(10, time.Minute), middlewares.AuthMiddleware).Post("/users/avatar", users.UploadAvatar)
		r.With(middlewares.RateLimitByIP(10, time.Minute), middlewares.AuthMiddleware).Post("/users/banner", users.UploadBanner)

		r.With(middlewares.RateLimitByIP(30, time.Minute), middlewares.AuthMiddleware).Get("/auth/me", auth.Me)

		r.With(middlewares.RateLimitByIP(30, time.Minute)).Get("/users/{username}", public.GetProfileByUsername)
		r.With(middlewares.RateLimitByIP(30, time.Minute)).Get("/users/{username}/posts", public.GetPostsByUsername)

		r.With(middlewares.RateLimitByIP(30, time.Minute), middlewares.AuthMiddleware).Get("/users/search", users.SearchUsers)

		r.With(middlewares.RateLimitByIP(5, time.Minute), middlewares.AuthMiddleware).Post("/posts", post.CreatePost)

		r.With(middlewares.RateLimitByIP(10, time.Minute), middlewares.AuthMiddleware).Post("/users/{username}/follow", users.FollowUser)
		r.With(middlewares.RateLimitByIP(10, time.Minute), middlewares.AuthMiddleware).Delete("/users/{username}/follow", users.UnfollowUser)
		r.With(middlewares.RateLimitByIP(30, time.Minute), middlewares.AuthMiddleware).Get("/users/{username}/follow", users.FollowState)
	})

	log.Println("[INFO] Server running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("[ERROR] server error: ", err)
	}
}
