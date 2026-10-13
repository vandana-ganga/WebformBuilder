package submission

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"webform-builder/internal/form"
)

// FormProvider provides access to form and published-version information.
type FormProvider interface {
	GetFormByID(ctx context.Context, id string) (*form.Form, error)
	GetPublishedFormVersion(ctx context.Context, formID string) (*form.FormVersion, error)
}

// Service contains submission business logic.
type Service struct {
	Repo         *Repository
	Publisher    *Publisher
	FormProvider FormProvider
}

// NewService creates a new submission service.
func NewService(
	repo *Repository,
	publisher *Publisher,
	formProvider FormProvider,
) *Service {
	return &Service{
		Repo:         repo,
		Publisher:    publisher,
		FormProvider: formProvider,
	}
}

// Create creates a new submission and publishes it to NATS JetStream.
//
// The API does not directly write the submission to PostgreSQL.
// Instead, it publishes the submission event to JetStream.
// The worker consumes the event and stores it in PostgreSQL.
func (s *Service)SubmitForm(
	ctx context.Context,
	formID string,
	data any,
) (*Submission, error) {

	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	if data == nil {
		return nil, errors.New("submission data is required")
	}

	// Get the form.
	currentForm, err := s.FormProvider.GetFormByID(ctx, formID)
	if err != nil {
		return nil, errors.New("form not found")
	}

	// Get the currently published version.
	publishedVersion, err := s.FormProvider.GetPublishedFormVersion(ctx, formID)
	if err != nil {
		return nil, errors.New("form has no published version")
	}

	// Make sure the published version belongs to this form.
	if publishedVersion.FormID != currentForm.ID {
		return nil, errors.New("published version does not belong to form")
	}

	// Validate the data against the published schema.
	if err := ValidateData(publishedVersion.Schema, data); err != nil {
		return nil, err
	}

	// Generate a unique submission ID.
	submissionID := uuid.New().String()

	createdAt := time.Now().UTC()

	// Create the event that will be sent to JetStream.
	message := SubmissionMessage{
		SubmissionID:  submissionID,
		TenantID:      currentForm.TenantID,
		FormID:        currentForm.ID,
		FormVersionID: publishedVersion.ID,
		Data:          data,
		CreatedAt:     createdAt,
	}

	// Publish to JetStream.
	// The worker will later persist this submission in PostgreSQL.
	if err := s.Publisher.Publish(ctx, message); err != nil {
		return nil, err
	}

	// Return the accepted submission information.
	return &Submission{
		ID:            submissionID,
		TenantID:      currentForm.TenantID,
		FormID:        currentForm.ID,
		FormVersionID: publishedVersion.ID,
		Data:          data,
		CreatedAt:     createdAt,
	}, nil
}

// GetSubmissionByID retrieves a submission by its ID.
func (s *Service) GetSubmissionByID(
	ctx context.Context,
	id string,
) (*Submission, error) {

	id = strings.TrimSpace(id)

	if id == "" {
		return nil, errors.New("submission ID is required")
	}

	return s.Repo.GetSubmissionByID(ctx, id)
}

// ListSubmissionsByForm returns submissions belonging to a form.
func (s *Service) ListSubmissionsByForm(
	ctx context.Context,
	formID string,
	limit int,
	offset int,
) ([]Submission, error) {

	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	// Default page size.
	if limit <= 0 {
		limit = 20
	}

	// Prevent very large queries.
	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.Repo.ListSubmissionsByForm(
		ctx,
		formID,
		limit,
		offset,
	)
}
