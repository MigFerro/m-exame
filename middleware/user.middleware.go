package middleware

import (
	"context"

	"github.com/MigFerro/exame/entities"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func WithAuthenticatedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		session, _ := session.Get("auth-cookie", c)
		val := session.Values["logged-user"]

		if _, ok := val.(*entities.AuthUser); ok {
			ctx := context.WithValue(c.Request().Context(), "authUser", val)
			c.SetRequest(c.Request().WithContext(ctx))
		}

		return next(c)
	}
}
