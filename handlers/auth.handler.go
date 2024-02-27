package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/local"
	"github.com/MigFerro/exame/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	UserService *services.UserService
}

func (h *AuthHandler) GetAuthProvider(c echo.Context) error {
	p := c.Param("provider")
	ctx := context.WithValue(c.Request().Context(), gothic.ProviderParamKey, p)

	_, err := gothic.CompleteUserAuth(c.Response(), c.Request().WithContext(ctx))
	if err == nil {
		fmt.Println("user already logged in")
		h.Login(c)
	}

	gothic.BeginAuthHandler(c.Response(), c.Request().WithContext(ctx))

	return nil
}

func (h *AuthHandler) Login(c echo.Context) error {

	// get provider from path params
	p := c.Param("provider")
	ctx := context.WithValue(c.Request().Context(), gothic.ProviderParamKey, p)

	gothUser, err := gothic.CompleteUserAuth(c.Response(), c.Request().WithContext(ctx))
	if err != nil {
		fmt.Println(err)
		return err
	}

	userExistsInDB := h.UserService.UserExistsInDB(gothUser.UserID)

	var dbUserId uuid.UUID

	if !userExistsInDB {

		dbUserId, err = h.UserService.CreateUserFromGoth(gothUser)

		if err != nil {
			fmt.Println("error creating user in DB after login. ", err)
			return err
		}
	}

	loggedUser := data.LoggedUser{
		Id:     dbUserId,
		AuthId: gothUser.UserID,
		Email:  gothUser.Email,
		Name:   gothUser.Name,
	}

	err = local.SaveLoggedUser(loggedUser, c, ctx)

	if err != nil {
		fmt.Println("error saving session: ", err)
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	local.RemoveLoggedUser(c)

	return c.Redirect(http.StatusFound, "/")
}
