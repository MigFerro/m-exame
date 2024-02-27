package local

import (
	"context"
	"fmt"

	"github.com/MigFerro/exame/data"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func SaveLoggedUser(user data.LoggedUser, echoContext echo.Context, context context.Context) error {
	// save user cookie
	session, _ := session.Get("session", echoContext)
	session.Values["logged-user"] = user
	err := session.Save(echoContext.Request().WithContext(context), echoContext.Response())

	return err
}

func RemoveLoggedUser(echoContext echo.Context) {
	session, _ := session.Get("session", echoContext)

	val := session.Values["logged-user"]
	if _, ok := val.(*data.LoggedUser); ok {
		session.Values["logged-user"] = nil
	} else {
		fmt.Println("Could not retrieve logged user from cookies.")
	}

	session.Save(echoContext.Request(), echoContext.Response())
}

func GetLoggedUser(echoContext echo.Context) (*data.LoggedUser, bool) {
	session, _ := session.Get("session", echoContext)

	val := session.Values["logged-user"]

	if val != nil {
		if _, ok := val.(*data.LoggedUser); ok {
			return val.(*data.LoggedUser), ok
		}
	}

	return nil, false
}

func IsLoggedIn(loggedUser *data.LoggedUser) bool {
	return loggedUser.Id == data.LoggedUser{}.Id
}
