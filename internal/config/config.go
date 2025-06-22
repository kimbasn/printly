package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   			string
	DBDriver 			string // "sqlite", "postgres", etc.
	DBSource 			string // DSN or file path
	Host				string
	Port     			string
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, loading from system ENV only")
	}

	cfg := &Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		DBDriver: getEnv("DB_DRIVER", "sqlite"),
		DBSource: getEnv("DB_SOURCE", "printly.db"),
		Host: getEnv("SERVER_ADDRESS", "localhost"),
		Port:     getEnv("PORT", "8080"),
	}

	return cfg
}
