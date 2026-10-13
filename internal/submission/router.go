package submission

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	"webform-builder/internal/form"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	js nats.JetStreamContext,
) {
	submissionRepo := NewRepository(db)
	formRepo := form.NewRepository(db)
	publisher := NewPublisher(js)
	service := NewService(submissionRepo, publisher, formRepo,)
	handler := NewHandler(service)

	e.POST("/public/forms/:id/submissions", handler.SubmitForm)
	e.GET("/api/submissions/:id", handler.GetSubmissionByID,)
	e.GET("/api/forms/:id/submissions", handler.ListSubmissionsByForm,)
}