package tenant

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailTaken = errors.New("email is already in use")

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
    return &Repository{
        DB: db,
    }
}

func (r *Repository)Create(ctx context.Context, name string, email string, passwordHash string,) (*Tenant, error) {
    var tenant Tenant

	query := `
		INSERT INTO tenants (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		name,
		email,
		passwordHash,
	).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Email,
		&tenant.CreatedAt,
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrEmailTaken
	}

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (r *Repository) GetByID(ctx context.Context, id string,) (*Tenant, error) {
	var tenant Tenant

	query := `
		SELECT id, name, COALESCE(email, ''), created_at
		FROM tenants
		WHERE id = $1
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Email,
		&tenant.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (r *Repository) Count(ctx context.Context,) (int, error) {
	var total int

	query := `
		SELECT COUNT(*)
		FROM tenants
	`

	err := r.DB.QueryRow(
		ctx,
		query,
	).Scan(&total)

	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *Repository) List(ctx context.Context, limit int, offset int,) ([]Tenant, error) {

	query := `
		SELECT id, name, COALESCE(email, ''), created_at
		FROM tenants
		ORDER BY created_at DESC, id
		LIMIT $1
		OFFSET $2
	`

	rows, err := r.DB.Query(
		ctx,
		query,
		limit,
		offset,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tenants := make([]Tenant, 0)

	for rows.Next() {

		var tenant Tenant

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Email,
			&tenant.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		tenants = append(
			tenants,
			tenant,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tenants, nil
}