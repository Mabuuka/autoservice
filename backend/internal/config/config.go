package config

import "os"

type Config struct {
	AppPort      string
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	DBSSLMode    string
	StaticDir    string
	TemplatesDir string
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func Load() Config {
	return Config{
		AppPort:      getEnv("APP_PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "db"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBName:       getEnv("DB_NAME", "automaster"),
		DBUser:       getEnv("DB_USER", "automaster_user"),
		DBPassword:   getEnv("DB_PASSWORD", "automaster_password"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		StaticDir:    getEnv("STATIC_DIR", "/app/frontend/static"),
		TemplatesDir: getEnv("TEMPLATES_DIR", "/app/frontend/templates"),
	}
}
