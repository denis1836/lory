package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"lory/internal/api/routes"
	"lory/internal/config"
	"lory/internal/db"
)

func main() {
	log.Println("Launching lory server...")
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	Pool, err := db.NewPool(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer func() {
		log.Println("Closing db pool...")
		Pool.Close()
	}()
	log.Println("Successfully connected to db!")

	router := routes.InitRoutes(Pool)

	log.Printf("Server is listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Printf("Server crashed: %v", err)
		os.Exit(1)
	}
}
