package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL     string
	RedisURL        string
	PaymentDeadline int // in minutes
}

func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/ticket_booking?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
		PaymentDeadline: getEnvAsInt("PAYMENT_DEADLINE", 15),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Handle DATABASE_URL construction from individual components
	if key == "DATABASE_URL" && value == "" {
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		dbUser := getEnv("DB_USER", "postgres")
		dbPassword := getEnv("DB_PASSWORD", "password")
		dbName := getEnv("DB_NAME", "ticket_booking")

		return strings.Join([]string{
			"postgres://",
			dbUser, ":",
			dbPassword, "@",
			dbHost, ":",
			dbPort, "/",
			dbName, "?sslmode=disable",
		}, "")
	}

	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
