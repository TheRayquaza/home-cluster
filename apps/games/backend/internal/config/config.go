package config

import "os"

type Config struct {
	DBURL            string
	JWTSecret        string
	Port             string
	AllowedOrigin    string
	BaseURL          string
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
}

func Load() Config {
	return Config{
		DBURL:            getEnv("DB_URL", "postgres://games:games@localhost:5432/games?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me"),
		Port:             getEnv("PORT", "8080"),
		AllowedOrigin:    getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		BaseURL:          getEnv("BASE_URL", "https://games.internal.rayq.app"),
		OIDCIssuer:       getEnv("OIDC_ISSUER", ""),
		OIDCClientID:     getEnv("OIDC_CLIENT_ID", "games"),
		OIDCClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  getEnv("OIDC_REDIRECT_URL", "https://games.internal.rayq.app/api/auth/oidc/callback"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
