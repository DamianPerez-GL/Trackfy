package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	FyEngine   FyEngineConfig
	FyAnalysis FyAnalysisConfig
}

type FyAnalysisConfig struct {
	URL     string
	Timeout time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL      string
	MaxConns int
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type FyEngineConfig struct {
	URL     string
	Timeout time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://trackfy:trackfy@localhost:5432/trackfy_gateway?sslmode=disable"),
			MaxConns: getIntEnv("DB_MAX_CONNS", 20),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:  getDurationEnv("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getDurationEnv("JWT_REFRESH_TTL", 7*24*time.Hour),
		},
		FyEngine: FyEngineConfig{
			URL:     getEnv("FY_ENGINE_URL", "http://fy-engine:8082"),
			Timeout: getDurationEnv("FY_ENGINE_TIMEOUT", 30*time.Second),
		},
		FyAnalysis: FyAnalysisConfig{
			URL:     getEnv("FY_ANALYSIS_URL", "http://fy-analysis:9090"),
			Timeout: getDurationEnv("FY_ANALYSIS_TIMEOUT", 30*time.Second),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
