package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerHost string
	ServerPort string
	Env        string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	FileStoragePath string

	SimilarityThreshold    float64
	MinTokensForComparison int
}

func LoadConfig() Config {
	threshold, _ := strconv.ParseFloat(getEnv("SIMILARITY_THRESHOLD", "0.85"), 64)
	minTokens, _ := strconv.Atoi(getEnv("MIN_TOKENS_FOR_COMPARISON", "50"))

	return Config{
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		Env:        getEnv("ENVIRONMENT", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "antiplague_user"),
		DBPassword: getEnv("DB_PASSWORD", "antiplague_password"),
		DBName:     getEnv("DB_NAME", "antiplague_db"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		FileStoragePath: getEnv("FILE_STORAGE_PATH", "./storage/files"),

		SimilarityThreshold:    threshold,
		MinTokensForComparison: minTokens,
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
