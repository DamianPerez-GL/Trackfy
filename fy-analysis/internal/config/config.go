package config

import (
	"os"
	"strconv"
)

// Config contiene la configuración de la aplicación
type Config struct {
	Port        string
	Environment string
	LogLevel    string
	RateLimit   int
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		RateLimit:   getEnvAsInt("RATE_LIMIT", 100),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
