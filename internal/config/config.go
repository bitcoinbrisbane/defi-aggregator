package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all our environment configurations
type Config struct {
	Port          string
	RedisURL      string
	NodeURL       string
	RedisPassword string
	APIKey        string
}

// Global config instance
var AppConfig Config

// InitConfig loads the configuration from environment variables or defaults
func InitConfig() Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize config with environment variables
	AppConfig = Config{
		Port:          GetEnvWithDefault("PORT", "8080"),
		RedisURL:      GetEnvWithDefault("REDIS_URL", "localhost:6379"),
		NodeURL:       GetEnvWithDefault("NODE_URL", "https://rpc-devnet.monadinfra.com/rpc/3fe540e310bbb6ef0b9f16cd23073b0a"),
		// NodeURL:       GetEnvWithDefault("NODE_URL", "https://eth-mainnet.g.alchemy.com/v2/fmiJslJk8E60f0Ni9QLq5nsnjm-lUzn1"),
		RedisPassword: GetEnvWithDefault("REDIS_PASSWORD", "Test1234"),
		APIKey:        GetEnvWithDefault("API_KEY", "your-api-key"),
	}

	log.Printf("Config loaded. Port: %s", AppConfig.Port)
	return AppConfig
}

// GetEnvWithDefault gets an environment variable or returns a default value
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetConfig returns the current configuration
func GetConfig() Config {
	return AppConfig
}