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
	exerciseHandler := handlers.ExerciseHandler{ExerciseService: &exerciseService, UsersService: &userService}
	authHandler := handlers.AuthHandler{UserService: &userService}

	app.GET("/", homeHandler.HomeShow)
	app.GET("/exam-list", homeHandler.ExameExerciseListShow)
	app.GET("/year-category-list", homeHandler.YearExerciseCategoryListShow)

	// auth
	app.GET("/auth/:provider", authHandler.GetAuthProvider)
	app.GET("/auth/:provider/callback", authHandler.Login)
	app.GET("/auth/logout", authHandler.Logout)

	//exercises
	app.GET("/exercises", exerciseHandler.ShowExerciseList)
	app.GET("/exercises/create", exerciseHandler.ShowExerciseCreate)
	app.GET("/exercises/:id/update", exerciseHandler.ShowExerciseUpdate)
	app.GET("/exercises/:id", exerciseHandler.ShowExerciseDetail)
	app.GET("/exercises/:id/choices", exerciseHandler.ShowExerciseChoices)
	app.GET("/exercises/categories", exerciseHandler.ShowExerciseCategoriesList)
	app.GET("/exercises/category/:id", exerciseHandler.ShowExerciseCategoryDetail)
	app.GET("/exercises/category/create", exerciseHandler.ShowCreateExerciseCategory)
	app.GET("/exercises/category/:id/edit", exerciseHandler.ShowUpdateExerciseCategory)
	app.GET("/exercises/history", exerciseHandler.ShowExerciseHistory)

	//test
	app.GET("/test", exerciseHandler.ShowTest)

	app.POST("/exercises/create", exerciseHandler.HandleExerciseUpsertJourney)
	app.POST("/exercises/:id/update", exerciseHandler.HandleExerciseUpsertJourney)
	app.POST("/exercises/:id/solve", exerciseHandler.HandleExerciseSolve)
	app.POST("/exercises/category/create", exerciseHandler.CreateExerciseCategory)
	app.POST("/exercises/category/:id", exerciseHandler.UpdateExerciseCategory)

	app.Start(":3000")
}
