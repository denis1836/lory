package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/db"
	"lory/internal/middleware"
)

func main() {
	log.Println("Launching lory server...")
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config %v", err)
	}

	Pool, err := db.NewPool(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal("failed to create db pool: %w", err)
	}
	defer func() {
		log.Println("Closing db pool...")
		Pool.Close()
	}()
	log.Println("Successfully connected to db!")

	router := api.NewRouter()
	middleware := middleware.Authenticate(Pool, cfg)

	//TODO: routes loading

	log.Printf("Server is listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Printf("Server crashed: %v", err)
		os.Exit(1)
	}
}
