package config

import "os"

type Config struct {
	DBURL         string
	JWTSecret     string
	Port          string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		DBURL:         getEnv("DB_URL", "postgres://games:games@localhost:5432/games?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me"),
		Port:          getEnv("PORT", "8080"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
