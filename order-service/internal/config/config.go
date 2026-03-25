package config

import "os"

type Config struct {
	TemporalAddress   string
	TemporalNamespace string
	HealthPort        string
	GatewayURL        string
}

func Load() Config {
	return Config{
		TemporalAddress:   getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		HealthPort:        getEnv("HEALTH_PORT", "8081"),
		GatewayURL:        getEnv("GATEWAY_URL", "http://temporal-gateway:8090"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
