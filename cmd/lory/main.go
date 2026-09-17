package main

import (
	"log"

	"lory/internal/config"
)

func main() {
	log.Println("Launching lory server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config %v", err)
	}

}
