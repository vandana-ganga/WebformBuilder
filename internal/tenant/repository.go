package tenant

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

func (r *Repository) Create(ctx context.Context, name string,) (*Tenant, error) {
    var tenant Tenant

	query := `
		INSERT INTO tenants (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		name,
	).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (r *Repository) GetByID(ctx context.Context, id string,) (*Tenant, error) {
	var tenant Tenant

	query := `
		SELECT id, name, created_at
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
		&tenant.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (r *Repository) List(ctx context.Context,) ([]Tenant, error) {

	query := `
		SELECT id, name, created_at
		FROM tenants
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(
		ctx,
		query,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tenants []Tenant

	for rows.Next() {

		var tenant Tenant

		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
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