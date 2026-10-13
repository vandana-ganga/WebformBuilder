package submission

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"webform-builder/internal/form"
)

type FormProvider interface {
	GetFormByID(ctx context.Context, id string) (*form.Form, error)
	GetPublishedFormVersion(ctx context.Context, formID string) (*form.FormVersion, error)
}

type Service struct {
	Repo         *Repository
	Publisher    *Publisher
	FormProvider FormProvider
}

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

	currentForm, err := s.FormProvider.GetFormByID(ctx, formID)
	if err != nil {
		return nil, errors.New("form not found")
	}

	publishedVersion, err := s.FormProvider.GetPublishedFormVersion(ctx, formID)
	if err != nil {
		return nil, errors.New("form has no published version")
	}

	if publishedVersion.FormID != currentForm.ID {
		return nil, errors.New("published version does not belong to form")
	}
	if err := ValidateData(publishedVersion.Schema, data); err != nil {
		return nil, err
	}

	submissionID := uuid.New().String()

	createdAt := time.Now().UTC()

	message := SubmissionMessage{
		SubmissionID:  submissionID,
		TenantID:      currentForm.TenantID,
		FormID:        currentForm.ID,
		FormVersionID: publishedVersion.ID,
		Data:          data,
		CreatedAt:     createdAt,
	}

	if err := s.Publisher.Publish(ctx, message); err != nil {
		return nil, err
	}

	return &Submission{
		ID:            submissionID,
		TenantID:      currentForm.TenantID,
		FormID:        currentForm.ID,
		FormVersionID: publishedVersion.ID,
		Data:          data,
		CreatedAt:     createdAt,
	}, nil
}

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


	if limit <= 0 {
		limit = 20
	}

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
