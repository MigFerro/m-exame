package handlers

import (
	"errors"

	"github.com/MigFerro/exame/services"
	"github.com/MigFerro/exame/templates/components"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	UsersService *services.UserService
}

func (h *UserHandler) GetLoggedUserPoints(c echo.Context) error {
	authUser, ok := getAuthenticatedUser(c.Request().Context())

	if !ok {
		return errors.New("No user authenticated.")
	}

	points, ok := h.UsersService.GetUserPoints(authUser.Id)

	if !ok {
		return errors.New("User is not student")
	}

	return render(c, components.HeaderUserPoints(points))

}
