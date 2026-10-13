package form

import (
	"net/http"

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

func (h *Handler)ListForms(c echo.Context) error {

	tenantID := c.Param("tenant_id")

	forms, err := h.Service.ListForms(
		c.Request().Context(),
		tenantID,
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

	versions, err := h.Service.ListFormVersions(
		c.Request().Context(),
		formID,
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