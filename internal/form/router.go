package form

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
) *echo.Echo {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	e.POST("/api/tenants/:tenant_id/forms", handler.CreateForm,)
	e.GET("/api/tenants/:tenant_id/forms", handler.ListForms,)
	e.POST("/api/forms/:id/versions", handler.CreateFormVersion,)
	e.GET("/api/forms/:id/versions", handler.ListFormVersions,)
	e.POST("/api/forms/:id/versions/:version_id/publish", handler.PublishFormVersion,)
	e.GET("/public/forms/:id", handler.GetPublishedFormVersion,)

	return e
}