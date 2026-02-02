package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	GinMode             string
	DatabaseURL         string
	JWTSecret           string
	JWTExpirationHours  int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExpHours = 24
	}

	return &Config{
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/agenda?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpirationHours: jwtExpHours,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
