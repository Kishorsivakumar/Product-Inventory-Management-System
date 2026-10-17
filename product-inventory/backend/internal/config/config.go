package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	AllowedOrigin     string
	LowStockThreshold int
}

func Load() Config {
	return Config{
		Port:              get("PORT", "8080"),
		DatabaseURL:       get("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable"),
		AllowedOrigin:     get("ALLOWED_ORIGIN", "http://localhost:4200"),
		LowStockThreshold: getInt("LOW_STOCK_THRESHOLD", 10),
	}
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}
