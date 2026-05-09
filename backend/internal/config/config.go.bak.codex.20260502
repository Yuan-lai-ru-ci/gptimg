package config

import (
	"os"
)

type Config struct {
	ServerPort     string
	DatabasePath   string
	JWTSecret      string
	EncryptionKey  string
	StoragePath    string
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabasePath:   getEnv("DATABASE_PATH", "./data/gptimg.db"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", "your-32-byte-encryption-key-here"),
		StoragePath:    getEnv("STORAGE_PATH", "./storage/images"),
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:3001"},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
