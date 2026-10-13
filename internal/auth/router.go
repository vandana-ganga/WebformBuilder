package auth

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	tokens *TokenManager,
	authMW *Middleware,
	adminEmail string,
	adminPassword string,
) {
	repo := NewRepository(db)
	service := NewService(repo, tokens, adminEmail, adminPassword)
	handler := NewHandler(service)

	e.POST("/api/auth/admin/login", handler.AdminLogin)
	e.POST("/api/auth/tenant/login", handler.TenantLogin)
	e.GET("/api/auth/me", handler.Me, authMW.RequireAuth)
}
