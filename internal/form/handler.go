package form

import (
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

type CreateFormRequest struct {
	Name string `json:"name"`
}

func (h *Handler)CreateForm(c echo.Context) error {
	tenantID := c.Param("tenant_id")
	var req CreateFormRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	form, err := h.Service.CreateForm(
		c.Request().Context(),
		tenantID,
		req.Name,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusCreated,
		form,
	)
}

func (h *Handler) GetFormByID(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "form ID is required",
		})
	}

	form, err := h.Service.GetFormDetail(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]any{
			"error": "form not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"form": form,
	})
}

func (h *Handler)ListFormsByTenant(c echo.Context) error {

	tenantID := c.Param("tenant_id")

	page := 1
	limit := 20

	if value := c.QueryParam("limit"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid limit",
				},
			)
		}

		limit = parsed
	}

	if value := c.QueryParam("page"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid page",
				},
			)
		}

		page = parsed
	}

	forms, err := h.Service.ListFormsByTenant(
		c.Request().Context(),
		tenantID,
		page,
		limit,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "failed to fetch forms",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		forms,
	)
}



type CreateVersionRequest struct {
	Schema any `json:"schema"`
}

func (h *Handler)CreateFormVersion(c echo.Context) error {

	formID := c.Param("id")

	var req CreateVersionRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	version, err := h.Service.CreateFormVersion(
		c.Request().Context(),
		formID,
		req.Schema,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusCreated,
		version,
	)
}
func (h *Handler) ListFormVersions(c echo.Context) error {

	formID := c.Param("id")

	page := 1
	limit := 20

	if value := c.QueryParam("limit"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid limit",
				},
			)
		}

		limit = parsed
	}

	if value := c.QueryParam("page"); value != "" {

		parsed, err := strconv.Atoi(value)

		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{
					"error": "invalid page",
				},
			)
		}

		page = parsed
	}

	versions, err := h.Service.ListFormVersions(
		c.Request().Context(),
		formID,
		page,
		limit,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "failed to fetch versions",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		versions,
	)
}

func (h *Handler) PublishFormVersion(c echo.Context) error {
	formID := c.Param("id")
	versionID := c.Param("version_id")

	version, err := h.Service.PublishFormVersion(
		c.Request().Context(),
		formID,
		versionID,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		version,
	)
}
func (h *Handler) GetPublishedFormVersion(c echo.Context) error {

	formID := c.Param("id")

	version, err := h.Service.GetPublishedFormVersion(
		c.Request().Context(),
		formID,
	)

	if err != nil {
		return c.JSON(
			http.StatusNotFound,
			map[string]string{
				"error": "published form not found",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		version,
	)
}