package form

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *Repository)CreateForm(ctx context.Context, tenantID string, name string,) (*Form, error) {
    var form Form
	query := `
		INSERT INTO forms (
			tenant_id,
			name,
			status
		)
		VALUES ($1, $2, 'draft')
		RETURNING
			id,
			tenant_id,
			name,
			status,
			created_at,
			updated_at
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		tenantID,
		name,
	).Scan(
		&form.ID,
		&form.TenantID,
		&form.Name,
		&form.Status,
		&form.CreatedAt,
		&form.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	return &form, nil
}

func (r *Repository)GetFormByID(ctx context.Context, id string,) (*Form, error) {
	var form Form
	query := `
		SELECT
			id,
			tenant_id,
			name,
			status,
			created_at,
			updated_at
		FROM forms
		WHERE id = $1
	`
	err := r.DB.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&form.ID,
		&form.TenantID,
		&form.Name,
		&form.Status,
		&form.CreatedAt,
		&form.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get form: %w", err)
	}

	return &form, nil
}

func (r *Repository)ListFormsByTenant(ctx context.Context, tenantID string,) ([]Form, error) {
	query := `
		SELECT
			id,
			tenant_id,
			name,
			status,
			created_at,
			updated_at
		FROM forms
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(
		ctx,
		query,
		tenantID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to list forms: %w",
			err,
		)
	}

	defer rows.Close()

	forms := make([]Form, 0)

	for rows.Next() {

		var form Form

		err := rows.Scan(
			&form.ID,
			&form.TenantID,
			&form.Name,
			&form.Status,
			&form.CreatedAt,
			&form.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan form: %w",
				err,
			)
		}

		forms = append(forms, form)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate forms: %w",
			err,
		)
	}

	return forms, nil
}

func (r *Repository)CreateFormVersion(ctx context.Context, formID string, schema any,) (*FormVersion, error) {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to encode schema: %w",
			err,
		)
	}

	tx, err := r.DB.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to begin transaction: %w",
			err,
		)
	}

	defer tx.Rollback(ctx)

	var formExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1
			FROM forms
			WHERE id = $1
		)
		`,
		formID,
	).Scan(&formExists)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to check form: %w",
			err,
		)
	}

	if !formExists {
		return nil, fmt.Errorf(
			"form not found",
		)
	}
	var versionNumber int

	err = tx.QueryRow(
		ctx,
		`
		SELECT COALESCE(
			MAX(version_number),
			0
		) + 1
		FROM form_versions
		WHERE form_id = $1
		`,
		formID,
	).Scan(&versionNumber)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to calculate version number: %w",
			err,
		)
	}

	var version FormVersion

	var storedSchema []byte

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO form_versions (
			form_id,
			version_number,
			schema
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			form_id,
			version_number,
			schema,
			created_at,
			published_at
		`,
		formID,
		versionNumber,
		schemaJSON,
	).Scan(
		&version.ID,
		&version.FormID,
		&version.VersionNumber,
		&storedSchema,
		&version.CreatedAt,
		&version.PublishedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create form version: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		storedSchema,
		&version.Schema,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode schema: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"failed to commit version: %w",
			err,
		)
	}

	return &version, nil
}

func (r *Repository)ListFormVersions(ctx context.Context, formID string, ) ([]FormVersion, error) {
	query := `
		SELECT
			id,
			form_id,
			version_number,
			schema,
			created_at,
			published_at
		FROM form_versions
		WHERE form_id = $1
		ORDER BY version_number DESC
	`

	rows, err := r.DB.Query(
		ctx,
		query,
		formID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to list versions: %w",
			err,
		)
	}
	defer rows.Close()

	versions := make([]FormVersion, 0)

	for rows.Next() {

		var version FormVersion

		var schemaJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.FormID,
			&version.VersionNumber,
			&schemaJSON,
			&version.CreatedAt,
			&version.PublishedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan version: %w",
				err,
			)
		}

		if err := json.Unmarshal(
			schemaJSON,
			&version.Schema,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to decode schema: %w",
				err,
			)
		}

		versions = append(
			versions,
			version,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate versions: %w",
			err,
		)
	}

	return versions, nil
}

func (r *Repository)GetFormVersionByID(ctx context.Context, versionID string,) (*FormVersion, error) {
	var version FormVersion
	var schemaJSON []byte

	query := `
		SELECT
			id,
			form_id,
			version_number,
			schema,
			created_at,
			published_at
		FROM form_versions
		WHERE id = $1
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		versionID,
	).Scan(
		&version.ID,
		&version.FormID,
		&version.VersionNumber,
		&schemaJSON,
		&version.CreatedAt,
		&version.PublishedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get version: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		schemaJSON,
		&version.Schema,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode schema: %w",
			err,
		)
	}

	return &version, nil
}

func (r *Repository)PublishFormVersion(ctx context.Context, formID string, versionID string,) (*FormVersion, error) {
	tx, err := r.DB.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to begin transaction: %w",
			err,
		)
	}

	defer tx.Rollback(ctx)

	var version FormVersion

	var schemaJSON []byte

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			id,
			form_id,
			version_number,
			schema,
			created_at,
			published_at
		FROM form_versions
		WHERE id = $1
		  AND form_id = $2
		`,
		versionID,
		formID,
	).Scan(
		&version.ID,
		&version.FormID,
		&version.VersionNumber,
		&schemaJSON,
		&version.CreatedAt,
		&version.PublishedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"form version not found: %w",
			err,
		)
	}

	err = tx.QueryRow(
		ctx,
		`
		UPDATE form_versions
		SET published_at = COALESCE(
			published_at,
			NOW()
		)
		WHERE id = $1
		RETURNING published_at
		`,
		versionID,
	).Scan(
		&version.PublishedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to publish version: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`
		UPDATE forms
		SET
			status = 'published',
			updated_at = NOW()
		WHERE id = $1
		`,
		formID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to update form status: %w",
			err,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"failed to commit publish transaction: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		schemaJSON,
		&version.Schema,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode schema: %w",
			err,
		)
	}

	return &version, nil
}
func (r *Repository)GetPublishedFormVersion(ctx context.Context, formID string,) (*FormVersion, error) {
	var version FormVersion
	var schemaJSON []byte

	query := `
		SELECT
			id,
			form_id,
			version_number,
			schema,
			created_at,
			published_at
		FROM form_versions
		WHERE form_id = $1
		  AND published_at IS NOT NULL
		ORDER BY version_number DESC
		LIMIT 1
	`

	err := r.DB.QueryRow(
		ctx,
		query,
		formID,
	).Scan(
		&version.ID,
		&version.FormID,
		&version.VersionNumber,
		&schemaJSON,
		&version.CreatedAt,
		&version.PublishedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"published form version not found: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		schemaJSON,
		&version.Schema,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode schema: %w",
			err,
		)
	}

	return &version, nil
}
