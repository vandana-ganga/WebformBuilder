package submission

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	"webform-builder/internal/auth"
	"webform-builder/internal/form"
)

func RegisterRoutes(
	e *echo.Echo,
	db *pgxpool.Pool,
	js nats.JetStreamContext,
	authMW *auth.Middleware,
) {
	submissionRepo := NewRepository(db)
	formRepo := form.NewRepository(db)
	publisher := NewPublisher(js)
	service := NewService(submissionRepo, publisher, formRepo,)
	handler := NewHandler(service)

	submissionAccess := authMW.RequireOwnerAccess(
		"id",
		"submission not found",
		func(ctx context.Context, id string) (string, error) {
			submission, err := service.GetSubmissionByID(ctx, id)
			if err != nil {
				return "", err
			}

			return submission.TenantID, nil
		},
	)

	formAccess := authMW.RequireOwnerAccess(
		"id",
		"form not found",
		func(ctx context.Context, id string) (string, error) {
			form, err := formRepo.GetFormByID(ctx, id)
			if err != nil {
				return "", err
			}

			return form.TenantID, nil
		},
	)

	e.POST("/public/forms/:id/submissions", handler.SubmitForm)
	e.GET("/api/submissions/:id", handler.GetSubmissionByID, submissionAccess, tenantAccess,)
	e.GET("/api/forms/:id/submissions", handler.ListSubmissionsByForm, formAccess, tenantAccess,)
}
