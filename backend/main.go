package main

import (
	"log"
	"net/http"

	"backend/internal/handlers/auth"
	"backend/internal/middlewars"
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

	if err := store.CreateTables(); err != nil {
		log.Fatal("[ERROR] create tables error: ", err)
	}

	router := chi.NewRouter()

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", auth.Register)
		r.Post("/auth/login", auth.Login)
		r.Post("/auth/refresh", auth.Refresh)

		r.With(middlewars.AuthMiddleware).Get("/auth/me", auth.Me)
	})

	log.Println("[INFO] Server running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("[ERROR] server error: ", err)
	}
}
