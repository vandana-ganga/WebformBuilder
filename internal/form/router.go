package form

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"webform-builder/internal/auth"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	authMW *auth.Middleware,
) *echo.Echo {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	tenantAccess := authMW.RequireTenantAccess("tenant_id")

	formAccess := authMW.RequireOwnerAccess(
		"id",
		"form not found",
		func(ctx context.Context, id string) (string, error) {
			form, err := service.GetFormByID(ctx, id)
			if err != nil {
				return "", err
			}

			return form.TenantID, nil
		},
	)

	e.POST("/api/tenants/:tenant_id/forms", handler.CreateForm, tenantAccess,)
	e.GET("/api/forms/:id", handler.GetFormByID, formAccess)
	e.GET("/api/tenants/:tenant_id/forms", handler.ListFormsByTenant, tenantAccess,)

	e.POST("/api/forms/:id/versions", handler.CreateFormVersion, formAccess)
	e.GET("/api/forms/:id/versions", handler.ListFormVersions, formAccess)
	e.POST("/api/forms/:id/versions/:version_id/publish", handler.PublishFormVersion, formAccess)
	e.GET("/public/forms/:id", handler.GetPublishedFormVersion,)

	return e
}
