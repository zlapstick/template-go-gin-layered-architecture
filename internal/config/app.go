package config

import (
	"log"
	"os"
	"strings"
)

type AppConfig struct {
	Port           string
	DBDsn          string
	JWTSecret      string
	AllowedOrigins []string
}

func AppLoad() *AppConfig {
	config := &AppConfig{
		Port:           os.Getenv("PORT"),
		DBDsn:          os.Getenv("DB_DSN"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		AllowedOrigins: parseOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.DBDsn == "" {
		log.Fatal("gagal memuat konfigurasi: DB_DSN is not set")
	}

	if config.JWTSecret == "" {
		log.Fatal("gagal memuat konfigurasi: JWT_SECRET is not set")
	}

	return config
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			origins = append(origins, part)
		}
	}

	if len(origins) == 0 {
		return []string{"*"}
	}

	return origins
}
