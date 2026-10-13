package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin  = "admin"
	RoleTenant = "tenant"
)

// Claims is the JWT payload.
// TenantID is only set for tenant tokens.
type Claims struct {
	Role     string `json:"role"`
	TenantID string `json:"tenant_id,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) IsAdmin() bool {
	return c.Role == RoleAdmin
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	Role      string    `json:"role"`
	TenantID  string    `json:"tenant_id,omitempty"`
}

// TenantCredentials is the login data stored on a tenant.
type TenantCredentials struct {
	TenantID     string
	PasswordHash string
}
