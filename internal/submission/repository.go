package submission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

func (r *Repository)SubmitForm(
	ctx context.Context,
	tenantID string,
	formID string,
	formVersionID string,
	submissionID string,
	data any,
	createdAt time.Time,
) (*Submission, error) {

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal submission data: %w",
			err,
		)
	}
	query := `
		INSERT INTO submissions (
			id,
			tenant_id,
			form_id,
			form_version_id,
			data,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
		RETURNING
			id,
			tenant_id,
			form_id,
			form_version_id,
			data,
			created_at
	`

	var submission Submission
	var storedData []byte

	err = r.DB.QueryRow(
		ctx,
		query,
		submissionID,
		tenantID,
		formID,
		formVersionID,
		dataJSON,
		createdAt,
	).Scan(
		&submission.ID,
		&submission.TenantID,
		&submission.FormID,
		&submission.FormVersionID,
		&storedData,
		&submission.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return r.GetSubmissionByID(ctx, submissionID)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create submission: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		storedData,
		&submission.Data,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode submission data: %w",
			err,
		)
	}

	return &submission, nil
}

func (r *Repository) GetSubmissionByID(
	ctx context.Context,
	id string,
) (*Submission, error) {

	query := `
		SELECT
			id,
			tenant_id,
			form_id,
			form_version_id,
			data,
			created_at
		FROM submissions
		WHERE id = $1
	`

	var submission Submission
	var dataJSON []byte

	err := r.DB.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&submission.ID,
		&submission.TenantID,
		&submission.FormID,
		&submission.FormVersionID,
		&dataJSON,
		&submission.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get submission: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		dataJSON,
		&submission.Data,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to decode submission data: %w",
			err,
		)
	}

	return &submission, nil
}

func (r *Repository) ListSubmissionsByForm(
	ctx context.Context,
	formID string,
	limit int,
	offset int,
) ([]Submission, error) {

	query := `
		SELECT
			id,
			tenant_id,
			form_id,
			form_version_id,
			data,
			created_at
		FROM submissions
		WHERE form_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.DB.Query(
		ctx,
		query,
		formID,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to list submissions: %w",
			err,
		)
	}

	defer rows.Close()

	submissions := make([]Submission, 0)

	for rows.Next() {

		var submission Submission
		var dataJSON []byte

		err := rows.Scan(
			&submission.ID,
			&submission.TenantID,
			&submission.FormID,
			&submission.FormVersionID,
			&dataJSON,
			&submission.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to scan submission: %w",
				err,
			)
		}

		if err := json.Unmarshal(
			dataJSON,
			&submission.Data,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to decode submission data: %w",
				err,
			)
		}

		submissions = append(
			submissions,
			submission,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed while reading submissions: %w",
			err,
		)
	}

	return submissions, nil
}