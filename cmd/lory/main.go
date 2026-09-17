package main

import (
	"context"
	"log"

	"lory/internal/config"
	"lory/internal/db"
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

}
