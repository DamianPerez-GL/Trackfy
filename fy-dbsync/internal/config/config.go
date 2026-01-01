package config

import "os"

// Config contiene la configuración del servicio
type Config struct {
	Port              string
	Environment       string
	LogLevel          string
	DatabaseURL       string
	PhishTankKey      string
	URLhausInterval   string
	PhishTankInterval string
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "9091"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		PhishTankKey:      getEnv("PHISHTANK_KEY", ""),
		URLhausInterval:   getEnv("URLHAUS_INTERVAL", "5m"),
		PhishTankInterval: getEnv("PHISHTANK_INTERVAL", "1h"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
