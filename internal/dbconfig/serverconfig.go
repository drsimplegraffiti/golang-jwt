package dbconfig

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	Environment string
	LogLevel    string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading file: %v", err)
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://admin@mydb--drsimplegraffiti.c-58.i.aws.edgedb.cloud:5656/main?sslmode=require"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
