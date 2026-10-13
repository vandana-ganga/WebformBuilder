package auth

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) GetTenantCredentialsByEmail(ctx context.Context, email string) (*TenantCredentials, error) {
	var creds TenantCredentials

	query := `
		SELECT id, password_hash
		FROM tenants
		WHERE LOWER(email) = LOWER($1)
		AND password_hash IS NOT NULL
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&creds.TenantID,
		&creds.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return &creds, nil
}
