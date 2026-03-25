package config

import "os"

type Config struct {
	Port            string
	TemporalAddress string
	DatabaseURL     string
	StaticDir       string
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8090"),
		TemporalAddress: getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://temporal:temporal@localhost:5432/workflow_designer?sslmode=disable"),
		StaticDir:       getEnv("STATIC_DIR", "./static"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
