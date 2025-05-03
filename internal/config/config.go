package config

import (
	"em_test/internal/logger"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	AppPort       string
	DB            DBConfig
	LogLevel      string
	EnrichmentAPI EnrichmentAPIConfig
}

type EnrichmentAPIConfig struct {
	AgifyURL       string
	GenderizeURL   string
	NationalizeURL string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
}

func Load() (*Config, error) {
	logger.Logger.Debug("Loading configuration")

	// Загружаем переменные окружения из файла .env
	if err := godotenv.Load(); err != nil {
		logger.Logger.Warn("Failed to load .env file", zap.Error(err))
		return nil, err
	}

	maxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "10"))

	config := &Config{
		AppPort:  getEnv("APP_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "people_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: int32(maxConns),
		},
		EnrichmentAPI: EnrichmentAPIConfig{
			AgifyURL:       getEnv("AGIFY_URL", "https://api.agify.io"),
			GenderizeURL:   getEnv("GENDERIZE_URL", "https://api.genderize.io"),
			NationalizeURL: getEnv("NATIONALIZE_URL", "https://api.nationalize.io"),
		},
	}

	logger.Logger.Info("Configuration loaded successfully", zap.Any("config", config))
	return config, nil
}

func getEnv(key, defaultValue string) string {
	logger.Logger.Debug("Getting environment variable", zap.String("key", key), zap.String("default", defaultValue))
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	logger.Logger.Info("Using default value for environment variable", zap.String("key", key), zap.String("default", defaultValue))
	return defaultValue
}
