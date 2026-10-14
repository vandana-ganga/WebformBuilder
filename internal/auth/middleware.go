package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const claimsContextKey = "auth_claims"

type Middleware struct {
	Tokens *TokenManager
}

func NewMiddleware(tokens *TokenManager) *Middleware {
	return &Middleware{
		Tokens: tokens,
	}
}

// RequireAuth accepts any valid token (admin or tenant)
// and stores its claims on the request context.
func (m *Middleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get(echo.HeaderAuthorization)

		scheme, token, found := strings.Cut(header, " ")

		if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			return unauthorized(c, "missing bearer token")
		}

		claims, err := m.Tokens.Parse(strings.TrimSpace(token))
		if err != nil {
			return unauthorized(c, err.Error())
		}

		c.Set(claimsContextKey, claims)

		return next(c)
	}
}

// RequireAdmin accepts only admin tokens.
func (m *Middleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return m.RequireAuth(func(c echo.Context) error {
		if !GetClaims(c).IsAdmin() {
			return Forbidden(c)
		}

		return next(c)
	})
}

// RequireTenantAccess accepts the admin, or the tenant whose ID
// matches the given path parameter.
func (m *Middleware) RequireTenantAccess(param string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return m.RequireAuth(func(c echo.Context) error {
			if !CanAccessTenant(c, c.Param(param)) {
				return Forbidden(c)
			}

			return next(c)
		})
	}
}

// OwnerLookup returns the tenant ID that owns the resource with the given ID.
type OwnerLookup func(ctx context.Context, id string) (string, error)

// RequireOwnerAccess accepts the admin, or the tenant that owns the
// resource named by the given path parameter. The owner is resolved
// with lookup; a failed lookup is reported as 404 with notFound.
func (m *Middleware) RequireOwnerAccess(param string, notFound string, lookup OwnerLookup) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return m.RequireAuth(func(c echo.Context) error {
			tenantID, err := lookup(
				c.Request().Context(),
				c.Param(param),
			)

			if err != nil {
				return c.JSON(
					http.StatusNotFound,
					map[string]string{
						"error": notFound,
					},
				)
			}

			if !CanAccessTenant(c, tenantID) {
				return Forbidden(c)
			}

			return next(c)
		})
	}
}

// GetClaims returns the claims set by RequireAuth, or nil.
func GetClaims(c echo.Context) *Claims {
	claims, _ := c.Get(claimsContextKey).(*Claims)

	return claims
}

// CanAccessTenant reports whether the caller is the admin
// or the tenant that owns the resource.
func CanAccessTenant(c echo.Context, tenantID string) bool {
	claims := GetClaims(c)

	if claims == nil {
		return false
	}

	if claims.IsAdmin() {
		return true
	}

	return tenantID != "" && strings.EqualFold(claims.TenantID, tenantID)
}

func Forbidden(c echo.Context) error {
	return c.JSON(
		http.StatusForbidden,
		map[string]string{
			"error": "forbidden",
		},
	)
}

func unauthorized(c echo.Context, message string) error {
	return c.JSON(
		http.StatusUnauthorized,
		map[string]string{
			"error": message,
		},
	)
}
