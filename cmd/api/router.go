package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	"webform-builder/internal/auth"
	"webform-builder/internal/config"
	"webform-builder/internal/form"
	"webform-builder/internal/submission"
	"webform-builder/internal/tenant"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	js nats.JetStreamContext,
	cfg *config.Config,
) {

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)
	authMW := auth.NewMiddleware(tokens)

	auth.RegisterRoutes(e, db, tokens, authMW, cfg.AdminEmail, cfg.AdminPassword)
	tenant.RegisterRoutes(e, db, authMW)
	form.RegisterRoutes(e, db, authMW)
	submission.RegisterRoutes(e, db, js, authMW)
}
