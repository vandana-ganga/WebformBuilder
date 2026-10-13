package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	NATSURL     string
	Port        string
}

func Load() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		NATSURL:     os.Getenv("NATS_URL"),
		Port:        port,
	}, nil
}