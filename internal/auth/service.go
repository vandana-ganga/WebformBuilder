package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Service struct {
	Repo          *Repository
	Tokens        *TokenManager
	AdminEmail    string
	AdminPassword string

	// Compared against when the tenant email is unknown, so that
	// login takes the same time whether or not the email exists.
	dummyPasswordHash string
}

func NewService(repo *Repository, tokens *TokenManager, adminEmail string, adminPassword string) *Service {
	dummyHash, _ := HashPassword("dummy-password")

	return &Service{
		Repo:              repo,
		Tokens:            tokens,
		AdminEmail:        adminEmail,
		AdminPassword:     adminPassword,
		dummyPasswordHash: dummyHash,
	}
}

// AdminLogin checks the single admin account configured in the environment.
func (s *Service) AdminLogin(email string, password string) (*LoginResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	emailOK := constantTimeEqual(email, strings.ToLower(s.AdminEmail))
	passwordOK := constantTimeEqual(password, s.AdminPassword)

	if !emailOK || !passwordOK {
		return nil, ErrInvalidCredentials
	}

	return s.issue(RoleAdmin, "")
}

func (s *Service) TenantLogin(ctx context.Context, email string, password string) (*LoginResponse, error) {
	email = strings.TrimSpace(email)

	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	creds, err := s.Repo.GetTenantCredentialsByEmail(ctx, email)

	if errors.Is(err, pgx.ErrNoRows) {
		CheckPassword(s.dummyPasswordHash, password)
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if !CheckPassword(creds.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return s.issue(RoleTenant, creds.TenantID)
}

func (s *Service) issue(role string, tenantID string) (*LoginResponse, error) {
	token, expiresAt, err := s.Tokens.Generate(role, tenantID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		Role:      role,
		TenantID:  tenantID,
	}, nil
}

func constantTimeEqual(a string, b string) bool {
	hashA := sha256.Sum256([]byte(a))
	hashB := sha256.Sum256([]byte(b))

	return subtle.ConstantTimeCompare(hashA[:], hashB[:]) == 1
}
