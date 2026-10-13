package submission

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		Service: service,
	}
}

type CreateSubmissionRequest struct {
	Data any `json:"data"`
}

func (h *Handler)SubmitForm(c echo.Context) error {
	ctx := c.Request().Context()

	formID := c.Param("id")

	if formID == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": "form ID is required",
			},
		)
	}

	var request CreateSubmissionRequest

	if err := c.Bind(&request); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": "invalid JSON request body",
			},
		)
	}

	if request.Data == nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": "data is required",
			},
		)
	}

	submission, err := h.Service.SubmitForm(
		ctx,
		formID,
		request.Data,
	)

	var validationErr *ValidationError

	if errors.As(err, &validationErr) {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error":  "validation failed",
				"fields": validationErr.Fields,
			},
		)
	}

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": err.Error(),
			},
		)
	}

	// 202 means:
	// "The request has been accepted for processing."
	//
	// The submission is currently in JetStream
	// and will be persisted by the worker.
	return c.JSON(
		http.StatusAccepted,
		map[string]any{
			"message":    "submission accepted",
			"submission": submission,
		},
	)
}

// GetSubmissionByID handles:
//
// GET /api/submissions/:id
func (h *Handler) GetSubmissionByID(c echo.Context) error {

	ctx := c.Request().Context()

	id := c.Param("id")

	if id == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": "submission ID is required",
			},
		)
	}

	submission, err := h.Service.GetSubmissionByID(
		ctx,
		id,
	)

	if err != nil {
		return c.JSON(
			http.StatusNotFound,
			map[string]any{
				"error": "submission not found",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		map[string]any{
			"submission": submission,
		},
	)
}

// ListByForm handles:
//
// GET /api/forms/:id/submissions
//
// Example:
//
// GET /api/forms/123/submissions?limit=20&offset=0
func (h *Handler) ListSubmissionsByForm(c echo.Context) error {

	ctx := c.Request().Context()

	formID := c.Param("id")

	if formID == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]any{
				"error": "form ID is required",
			},
		)
	}

	limit := 20
	offset := 0

	// Read limit query parameter.
	if value := c.QueryParam("limit"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{
					"error": "invalid limit",
				},
			)
		}

		limit = parsed
	}

	// Read offset query parameter.
	if value := c.QueryParam("offset"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{
					"error": "invalid offset",
				},
			)
		}

		offset = parsed
	}

	submissions, err := h.Service.ListSubmissionsByForm(
		ctx,
		formID,
		limit,
		offset,
	)

	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			map[string]any{
				"error": "failed to retrieve submissions",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		map[string]any{
			"form_id":     formID,
			"submissions": submissions,
			"limit":       limit,
			"offset":      offset,
		},
	)
}