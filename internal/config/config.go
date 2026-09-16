package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTExpirationHours int
}

func LoadConfig() *Config {
	// Attempt to load .env file if present, ignore if not found
	if err := godotenv.Load(); err != nil {
		log.Println("Notice: No .env file found or error loading, reading from system environment")
	}

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ticketdb?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "change-me")

	expHoursStr := getEnv("JWT_EXPIRATION_HOURS", "24")
	expHours, err := strconv.Atoi(expHoursStr)
	if err != nil {
		expHours = 24
	}

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		JWTExpirationHours: expHours,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
