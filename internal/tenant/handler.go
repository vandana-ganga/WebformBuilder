package tenant

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type CreateTenantRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{
        Service: service,
    }
}

func (h *Handler)TenantCreate(c echo.Context,) error {
	var req CreateTenantRequest

	err := c.Bind(&req)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	tenant, err := h.Service.Create(
		c.Request().Context(),
		req.Name,
		req.Email,
		req.Password,
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
		tenant,
	)
}

func (h *Handler)TenantGetByID(c echo.Context,) error {

	id := c.Param("id")

	tenant, err := h.Service.GetByID(
		c.Request().Context(),
		id,
	)

	if err != nil {
		return c.JSON( 
			http.StatusNotFound,
			map[string]string{
				"error": "tenant not found",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		tenant,
	)
}

func (h *Handler) ListTenants(c echo.Context,) error {

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

	tenants, err := h.Service.List(
		c.Request().Context(),
		page,
		limit,
	)

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "failed to get tenants",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		tenants,
	)
}