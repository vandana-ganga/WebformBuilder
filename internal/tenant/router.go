package tenant

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"webform-builder/internal/auth"
)

func RegisterRoutes(e *echo.Echo, db *pgxpool.Pool, authMW *auth.Middleware) {

	tenantRepo := NewRepository(db)
	tenantService := NewService(tenantRepo)
	tenantHandler := NewHandler(tenantService)

	e.POST("/api/tenants", tenantHandler.TenantCreate, authMW.RequireAdmin)
	e.GET("/api/tenants", tenantHandler.ListTenants, authMW.RequireAdmin)
	e.GET("/api/tenants/:id", tenantHandler.TenantGetByID, authMW.RequireTenantAccess("id"))
}