package tenant

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"webform-builder/internal/auth"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{
        Repo: repo,
    }
}

func (s *Service) Create(ctx context.Context, name string, email string, password string,) (*Tenant, error) {
    name = strings.TrimSpace(name)

	if name == "" {
		return nil, errors.New(
			"tenant name is required",
		)
	}

	if len(name) > 255 {
		return nil, errors.New(
			"tenant name is too long",
		)
	}

	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return nil, errors.New(
			"tenant email is required",
		)
	}

	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email || len(email) > 255 {
		return nil, errors.New(
			"tenant email is invalid",
		)
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return s.Repo.Create(
		ctx,
		name,
		email,
		passwordHash,
	)
}

func (s *Service) GetByID(ctx context.Context, id string,) (*Tenant, error) {
	if id == "" {
		return nil, errors.New(
			"tenant ID is required",
		)
	}
	return s.Repo.GetByID(
		ctx,
		id,
	)
}

func (s *Service) List(ctx context.Context, page int, limit int,) (*TenantList, error) {

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	total, err := s.Repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	tenants, err := s.Repo.List(
		ctx,
		limit,
		(page-1)*limit,
	)
	if err != nil {
		return nil, err
	}

	return &TenantList{
		Tenants: tenants,
		Page:    page,
		Limit:   limit,
		Total:   total,
	}, nil
}