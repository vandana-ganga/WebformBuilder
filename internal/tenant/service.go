package tenant

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

func (s *Service) Create(ctx context.Context, name string,) (*Tenant, error) {
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
	return s.Repo.Create(
		ctx,
		name,
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

func (s *Service) List(ctx context.Context,) ([]Tenant, error) {

	return s.Repo.List(ctx)
}