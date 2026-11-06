package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DSN         string
	WorkerCount int
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	return &Config{
		Port:        os.Getenv("APP_PORT"),
		DSN:         dsn,
		WorkerCount: getEnvAsInt("WORKER_COUNT", 5),
	}, nil
}
