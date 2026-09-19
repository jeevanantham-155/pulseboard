package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MongoDBURI      string
	MongoDBDatabase string
	RedisURL        string
	JWTSecret       string
	FrontendURL     string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	config := Config{
		Port:            getEnvOrDefault("PORT", "8080"),
		MongoDBURI:      os.Getenv("MONGODB_URI"),
		MongoDBDatabase: getEnvOrDefault("MONGODB_DATABASE", "live_polling"),
		RedisURL:        os.Getenv("REDIS_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		FrontendURL:     getEnvOrDefault("FRONTEND_URL", "http://localhost:5173"),
	}

	missing := make([]string, 0, 3)
	if config.MongoDBURI == "" {
		missing = append(missing, "MONGODB_URI")
	}
	if config.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if config.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return config, nil
}

func getEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
