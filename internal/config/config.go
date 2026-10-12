package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
}

func Load() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}, nil
}