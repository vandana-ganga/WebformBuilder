package form

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		Repo: repo,
	}
}

func (s *Service)CreateForm(ctx context.Context, tenantID string, name string,) (*Form, error) {
	tenantID = strings.TrimSpace(tenantID)
	name = strings.TrimSpace(name)

	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	if name == "" {
		return nil, errors.New("form name is required")
	}

	if len(name) > 255 {
		return nil, errors.New("form name is too long")
	}

	return s.Repo.CreateForm(
		ctx,
		tenantID,
		name,
	)
}

func (s *Service) GetFormById(ctx context.Context, id string,) (*Form, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("form ID is required")
	}

	return s.Repo.GetFormByID(
		ctx,
		id,
	)
}


func (s *Service) ListForms(ctx context.Context, tenantID string,) ([]Form, error) {
	tenantID = strings.TrimSpace(tenantID)

	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	return s.Repo.ListFormsByTenant(
		ctx,
		tenantID,
	)
}

func (s *Service) CreateFormVersion(ctx context.Context, formID string, schema any,) (*FormVersion, error) {

	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	if schema == nil {
		return nil, errors.New("schema is required")
	}

	return s.Repo.CreateFormVersion(
		ctx,
		formID,
		schema,
	)
}


func (s *Service) ListFormVersions(ctx context.Context, formID string,) ([]FormVersion, error) {

	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	return s.Repo.ListFormVersions(
		ctx,
		formID,
	)
}


func (s *Service) GetFormVersion( ctx context.Context, versionID string,) (*FormVersion, error) {
	versionID = strings.TrimSpace(versionID)
	if versionID == "" {
		return nil, errors.New("version ID is required")
	}
	return s.Repo.GetFormVersionByID(
		ctx,
		versionID,
	)
}


func (s *Service) PublishFormVersion(ctx context.Context, formID string, versionID string,) (*FormVersion, error) {
	formID = strings.TrimSpace(formID)
	versionID = strings.TrimSpace(versionID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	if versionID == "" {
		return nil, errors.New("version ID is required")
	}

	return s.Repo.PublishFormVersion(
		ctx,
		formID,
		versionID,
	)
}


func (s *Service) GetPublishedFormVersion(ctx context.Context, formID string,) (*FormVersion, error) {
	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	return s.Repo.GetPublishedFormVersion(
		ctx,
		formID,
	)
}