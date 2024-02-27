package middleware

import (
	"context"

	"github.com/MigFerro/exame/local"
	"github.com/labstack/echo/v4"
)

func WithAuthenticatedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		loggedUser, ok := local.GetLoggedUser(c)

		if ok {
			ctx := context.WithValue(c.Request().Context(), "authUser", loggedUser)
			c.SetRequest(c.Request().WithContext(ctx))
		}

		return next(c)
	}
}
