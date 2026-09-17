package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBUrl     string
	JWTSecret string
	SMTPPass  string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	// site
	port := getEnvOrDefault("PORT", "5432")

	// db
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPass == "" || dbName == "" {
		return nil, fmt.Errorf("missing .env fields (DB_USER, DB_PASS, DB_NAME)")
	}

	dbUrl := fmt.Sprintf("postgres://%s:%s/@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	// security
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("failed to load JWT_SECRET")
	}

	// smtp
	smtpPass := os.Getenv("SMTP_PASS")

	return &Config{
		Port:      port,
		DBUrl:     dbUrl,
		JWTSecret: jwtSecret,
		SMTPPass:  smtpPass,
	}, nil
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
