package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	NATSURL     string
	Port        string

	JWTSecret     string
	JWTExpiry     time.Duration
	AdminEmail    string
	AdminPassword string
}

func Load() (*Config, error) {

	// .env is optional: in Docker the settings come from the environment.
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	expiryHours := 24
	if value := os.Getenv("JWT_EXPIRY_HOURS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return nil, errors.New("JWT_EXPIRY_HOURS must be a positive number")
		}

		expiryHours = parsed
	}

	return &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		NATSURL:       os.Getenv("NATS_URL"),
		Port:          port,
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiry:     time.Duration(expiryHours) * time.Hour,
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}, nil
}

// ValidateAuth checks the settings the API needs for login.
// The worker does not use them, so Load does not require them.
func (c *Config) ValidateAuth() error {
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}

	if c.AdminEmail == "" || c.AdminPassword == "" {
		return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD are required")
	}

	return nil
}
