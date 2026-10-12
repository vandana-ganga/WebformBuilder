package tenant

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, db *pgxpool.Pool) {

	tenantRepo := NewRepository(db)
	tenantService := NewService(tenantRepo)
	tenantHandler := NewHandler(tenantService)

	e.POST("/api/tenants", tenantHandler.TenantCreate)
	e.GET("/api/tenants", tenantHandler.ListTenants)
	e.GET("/api/tenants/:id", tenantHandler.TenantGetByID)
}