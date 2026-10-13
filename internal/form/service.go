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

func (s *Service) GetFormByID(ctx context.Context, id string,) (*Form, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("form ID is required")
	}

	return s.Repo.GetFormByID(
		ctx,
		id,
	)
}


func (s *Service) GetFormDetail(ctx context.Context, id string,) (*FormDetail, error) {
	form, err := s.GetFormByID(ctx, id)
	if err != nil {
		return nil, err
	}

	versionIDs, err := s.Repo.ListFormVersionIDs(
		ctx,
		form.ID,
	)
	if err != nil {
		return nil, err
	}

	return &FormDetail{
		Form:         *form,
		VersionCount: len(versionIDs),
		VersionIDs:   versionIDs,
	}, nil
}

func (s *Service) ListFormsByTenant(ctx context.Context, tenantID string, page int, limit int,) (*FormList, error) {
	tenantID = strings.TrimSpace(tenantID)

	if tenantID == "" {
		return nil, errors.New("tenant ID is required")
	}

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	total, err := s.Repo.CountFormsByTenant(
		ctx,
		tenantID,
	)
	if err != nil {
		return nil, err
	}

	forms, err := s.Repo.ListFormsByTenant(
		ctx,
		tenantID,
		limit,
		(page-1)*limit,
	)
	if err != nil {
		return nil, err
	}

	return &FormList{
		Forms: forms,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
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


func (s *Service) ListFormVersions(ctx context.Context, formID string, page int, limit int,) (*FormVersionList, error) {

	formID = strings.TrimSpace(formID)

	if formID == "" {
		return nil, errors.New("form ID is required")
	}

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	total, err := s.Repo.CountFormVersions(
		ctx,
		formID,
	)
	if err != nil {
		return nil, err
	}

	versions, err := s.Repo.ListFormVersions(
		ctx,
		formID,
		limit,
		(page-1)*limit,
	)
	if err != nil {
		return nil, err
	}

	return &FormVersionList{
		Versions: versions,
		Page:     page,
		Limit:    limit,
		Total:    total,
	}, nil
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