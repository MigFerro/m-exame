package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/local"
	"github.com/MigFerro/exame/services"
	errorsview "github.com/MigFerro/exame/templates/errors"
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

	dbUser, userExistsInDB := h.UserService.UserExistsInDB(gothUser.UserID)

	if !userExistsInDB {

		dbUser, err = h.UserService.CreateUserFromGoth(gothUser)

		if err != nil {
			fmt.Println("error creating user in DB after login. ", err)
			return err
		}
	}

	loggedUser := data.LoggedUser{
		Id:     dbUser.Id,
		AuthId: gothUser.UserID,
		Email:  gothUser.Email,
		Name:   gothUser.Name,
	}

	err = local.SaveLoggedUser(loggedUser, c, ctx)

	if err != nil {
		fmt.Println("error saving session: ", err)
		return render(c, errorsview.GeneralErrorPage())
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	local.RemoveLoggedUser(c)

	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		return render(c, errorsview.GeneralErrorPage())
	}

	scheme := "http"
	if os.Getenv("IS_PROD") == "1" {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + c.Request().Host)
	if err != nil {
		return render(c, errorsview.GeneralErrorPage())
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}
