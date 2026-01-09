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

	// URL Engine Config
	GoogleWebRiskKey string
	URLScanKey       string
	PhishTankKey     string
	URLhausDBPath    string
	PhishTankDBPath  string
	EnableDBSync     bool

	// PostgreSQL Local DB
	DatabaseURL       string
	EnableLocalDB     bool
	EnableUserReports bool
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "9090"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		RateLimit:   getEnvAsInt("RATE_LIMIT", 100),

		// URL Engine - API Keys
		GoogleWebRiskKey: getEnv("GOOGLE_WEBRISK_KEY", ""),
		URLScanKey:       getEnv("URLSCAN_KEY", ""),
		PhishTankKey:     getEnv("PHISHTANK_KEY", ""),

		// URL Engine - DB Paths
		URLhausDBPath:   getEnv("URLHAUS_DB_PATH", "/data/urlhaus.csv"),
		PhishTankDBPath: getEnv("PHISHTANK_DB_PATH", "/data/phishtank.json"),
		EnableDBSync:    getEnvAsBool("ENABLE_DB_SYNC", true),

		// PostgreSQL Local DB
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		EnableLocalDB:     getEnvAsBool("ENABLE_LOCAL_DB", true),
		EnableUserReports: getEnvAsBool("ENABLE_USER_REPORTS", true),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
