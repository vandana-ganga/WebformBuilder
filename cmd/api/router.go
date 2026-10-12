package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"webform-builder/internal/tenant"
)

func RegisterRoutes(e *echo.Echo, db *pgxpool.Pool) {

	tenant.RegisterRoutes(e, db)
}