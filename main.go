package main

import (
	"github.com/MigFerro/exame/cookies"
	"github.com/MigFerro/exame/db"
	"github.com/MigFerro/exame/handlers"
	"github.com/MigFerro/exame/middleware"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func main() {

	sqlxDB := db.ConnectToDB()
	authCookieStore := cookies.InitCookieStore()

	app := echo.New()

	app.Use(session.Middleware(authCookieStore))
	app.Use(middleware.WithAuthenticatedUser)

	app.Static("/static", "templates/static")

	homeHandler := handlers.HomeHandler{}
	exerciseHandler := handlers.ExerciseHandler{DB: sqlxDB}
	authHandler := handlers.AuthHandler{DB: sqlxDB}

	app.GET("/", homeHandler.HomeShow)

	// auth
	app.GET("/auth/:provider", authHandler.GetAuthProvider)
	app.GET("/auth/:provider/callback", authHandler.Login)
	app.GET("/auth/logout", authHandler.Logout)

	//exercises
	app.GET("/exercises/:id", exerciseHandler.ExerciseShow)
	app.POST("/exercises/:id/solve", exerciseHandler.ExerciseSolve)

	app.Start(":3000")
}
