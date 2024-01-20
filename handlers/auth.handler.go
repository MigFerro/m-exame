package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MigFerro/exame/entities"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	DB *sqlx.DB
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

	// check if user exists in DB
	row := h.DB.QueryRowx("SELECT * FROM users where auth_id = $1", gothUser.UserID)

	var dbUser entities.UserEntity
	err = row.StructScan(&dbUser)

	if err != nil {

		// create user in database
		tx := h.DB.MustBegin()
		tx.MustExec("INSERT INTO users (auth_id, email, name) VALUES ($1, $2, $3)",
			gothUser.UserID,
			gothUser.Email,
			gothUser.Name,
		)
		row = tx.QueryRowx("SELECT * FROM users where auth_id = $1", gothUser.UserID)
		err = row.StructScan(&dbUser)
		tx.Commit()

		if err != nil {
			fmt.Println("Couldn't retrieve created user. ", err)
			return err
		}
	}

	// save user cookie
	session, _ := session.Get("auth-cookie", c)
	session.Values["logged-user"] = entities.AuthUser{
		Id:     dbUser.Id,
		AuthId: gothUser.UserID,
		Email:  gothUser.Email,
		Name:   gothUser.Name,
	}
	err = session.Save(c.Request().WithContext(ctx), c.Response())

	if err != nil {
		fmt.Println("error saving session: ", err)
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	session, _ := session.Get("auth-cookie", c)

	val := session.Values["logged-user"]
	if _, ok := val.(*entities.AuthUser); ok {
		fmt.Println("ok")
		session.Values["logged-user"] = entities.AuthUser{}
	} else {
		fmt.Println("Could not retrieve logged user from cookies.")
	}

	session.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/")
}
