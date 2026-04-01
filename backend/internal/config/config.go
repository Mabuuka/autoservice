package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort          string
	DBHost           string
	DBPort           string
	DBName           string
	DBUser           string
	DBPassword       string
	DBSSLMode        string
	StaticDir        string
	TemplatesDir     string
	AuthCookieName   string
	AuthSecret       string
	AuthCookieSecure bool
	AuthCookieTTL    int
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func Load() Config {
	return Config{
		AppPort:          getEnv("APP_PORT", "8080"),
		DBHost:           getEnv("DB_HOST", "db"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBName:           getEnv("DB_NAME", "automaster"),
		DBUser:           getEnv("DB_USER", "automaster_user"),
		DBPassword:       getEnv("DB_PASSWORD", "automaster_password"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		StaticDir:        getEnv("STATIC_DIR", "/app/frontend/static"),
		TemplatesDir:     getEnv("TEMPLATES_DIR", "/app/frontend/templates"),
		AuthCookieName:   getEnv("AUTH_COOKIE_NAME", "autoservice_auth"),
		AuthSecret:       getEnv("AUTH_SECRET", "dev-auth-secret-change-me"),
		AuthCookieSecure: getEnvBool("AUTH_COOKIE_SECURE", false),
		AuthCookieTTL:    getEnvInt("AUTH_COOKIE_TTL_HOURS", 72),
	}
}
