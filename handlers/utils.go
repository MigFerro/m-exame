package handlers

import (
	"context"
	"fmt"

	"github.com/MigFerro/exame/entities"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response())
}

func getAuthenticatedUser(c context.Context) *entities.AuthUser {
	authUser, ok := c.Value("authUser").(*entities.AuthUser)
	if !ok {
		fmt.Println("Could not get authenticated user.")
		return &entities.AuthUser{}
	}
	return authUser
}
