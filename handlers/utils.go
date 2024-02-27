package handlers

import (
	"context"
	"fmt"

	"github.com/MigFerro/exame/data"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response())
}

func getAuthenticatedUser(c context.Context) (*data.LoggedUser, bool) {
	authUser, ok := c.Value("authUser").(*data.LoggedUser)
	if !ok {
		fmt.Println("Could not get authenticated user.")
		return &data.LoggedUser{}, ok
	}
	return authUser, true
}
