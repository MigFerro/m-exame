package handlers

import (
	"fmt"

	"github.com/MigFerro/exame/services"
	"github.com/MigFerro/exame/templates/components"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	UsersService *services.UserService
}

func (h *UserHandler) GetLoggedUserPrepLevel(c echo.Context) error {
	authUser, ok := getAuthenticatedUser(c.Request().Context())

	if !ok {
		return render(c, components.HeaderUserPrepLevelFailed())
	}

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "student" {
		return render(c, components.HeaderUserPrepLevelFailed())
	}

	prepLevel, err := h.UsersService.GetPreparationLevel(authUser.Id)

	if err != nil {
		fmt.Println("error getting user preparation level", err)
		return render(c, components.HeaderUserPrepLevelFailed())
	}

	return render(c, components.HeaderUserPrepLevel(prepLevel))

}

func (h *UserHandler) GetLoggedUserPoints(c echo.Context) error {
	authUser, ok := getAuthenticatedUser(c.Request().Context())

	if !ok {
		return render(c, components.HeaderUserPointsFailed())
	}

	points, ok := h.UsersService.GetUserPoints(authUser.Id)

	if !ok {
		fmt.Println("error getting user points")
		return render(c, components.HeaderUserPointsFailed())
	}

	return render(c, components.HeaderUserPoints(points))

}
