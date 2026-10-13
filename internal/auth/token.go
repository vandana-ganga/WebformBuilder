package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

// TokenManager signs and verifies JWTs using HS256.
type TokenManager struct {
	Secret []byte
	Expiry time.Duration
}

func NewTokenManager(secret string, expiry time.Duration) *TokenManager {
	return &TokenManager{
		Secret: []byte(secret),
		Expiry: expiry,
	}
}

func (m *TokenManager) Generate(role string, tenantID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.Expiry)

	subject := role
	if tenantID != "" {
		subject = tenantID
	}

	claims := Claims{
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	).SignedString(m.Secret)

	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (m *TokenManager) Parse(tokenString string) (*Claims, error) {
	var claims Claims

	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (any, error) {
			return m.Secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	switch claims.Role {
	case RoleAdmin:
	case RoleTenant:
		if claims.TenantID == "" {
			return nil, ErrInvalidToken
		}
	default:
		return nil, ErrInvalidToken
	}

	return &claims, nil
}
