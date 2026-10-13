package auth

import (
	"errors"
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

func (h *Handler) AdminLogin(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	resp, err := h.Service.AdminLogin(
		req.Email,
		req.Password,
	)

	if err != nil {
		return loginError(c, err)
	}

	return c.JSON(
		http.StatusOK,
		resp,
	)
}

func (h *Handler) TenantLogin(c echo.Context) error {
	var req LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	resp, err := h.Service.TenantLogin(
		c.Request().Context(),
		req.Email,
		req.Password,
	)

	if err != nil {
		return loginError(c, err)
	}

	return c.JSON(
		http.StatusOK,
		resp,
	)
}

// Me returns the identity carried by the current token.
func (h *Handler) Me(c echo.Context) error {
	claims := GetClaims(c)

	return c.JSON(
		http.StatusOK,
		map[string]any{
			"role":       claims.Role,
			"tenant_id":  claims.TenantID,
			"expires_at": claims.ExpiresAt.Time,
		},
	)
}

func loginError(c echo.Context, err error) error {
	if errors.Is(err, ErrInvalidCredentials) {
		return c.JSON(
			http.StatusUnauthorized,
			map[string]string{
				"error": err.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusInternalServerError,
		map[string]string{
			"error": "login failed",
		},
	)
}
