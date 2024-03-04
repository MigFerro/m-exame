package main

import (
	"github.com/MigFerro/exame/cookies"
	"github.com/MigFerro/exame/db"
	"github.com/MigFerro/exame/handlers"
	"github.com/MigFerro/exame/middleware"
	"github.com/MigFerro/exame/services"
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

	exerciseService := services.ExerciseService{DB: sqlxDB}
	userService := services.UserService{DB: sqlxDB}

	homeHandler := handlers.HomeHandler{ExerciseService: &exerciseService}
	exerciseHandler := handlers.ExerciseHandler{ExerciseService: &exerciseService}
	authHandler := handlers.AuthHandler{UserService: &userService}

	app.GET("/", homeHandler.HomeShow)
	app.GET("/home/:id", homeHandler.HomeShow)

	// auth
	app.GET("/auth/:provider", authHandler.GetAuthProvider)
	app.GET("/auth/:provider/callback", authHandler.Login)
	app.GET("/auth/logout", authHandler.Logout)

	//exercises
	app.GET("/exercises", exerciseHandler.ShowExerciseList)
	app.GET("/exercises/create", exerciseHandler.ShowExerciseCreate)
	app.POST("/exercises/create", exerciseHandler.HandleExerciseUpsertJourney)
	app.GET("/exercises/:id/update", exerciseHandler.ShowExerciseUpdate)
	app.POST("/exercises/:id/update", exerciseHandler.HandleExerciseUpsertJourney)
	app.GET("/exercises/:id", exerciseHandler.ShowExerciseDetail)
	app.GET("/exercises/:id/home", exerciseHandler.ShowExerciseHomepage)
	app.POST("/exercises/:id/solve", exerciseHandler.HandleExerciseSolve)
	app.GET("/exercises/:id/choices", exerciseHandler.ShowExerciseChoices)
	app.GET("/exercises/categories", exerciseHandler.ShowExerciseCategoriesList)

	app.Start(":3000")
}
