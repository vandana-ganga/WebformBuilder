package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	// bcrypt ignores everything after 72 bytes.
	MaxPasswordLength = 72
)

func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", errors.New("password must be at least 8 characters")
	}

	if len(password) > MaxPasswordLength {
		return "", errors.New("password must be at most 72 characters")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)

	return err == nil
}
