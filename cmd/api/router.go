package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	"webform-builder/internal/form"
	"webform-builder/internal/submission"
	"webform-builder/internal/tenant"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	js nats.JetStreamContext,
) {

	tenant.RegisterRoutes(e, db,)
	form.RegisterRoutes(e, db,)
	submission.RegisterRoutes(e, db, js,)
}