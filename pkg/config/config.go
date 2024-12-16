package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig(path string) error {
	if err := godotenv.Load(path); err != nil {
		return err
	}

	return nil
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
