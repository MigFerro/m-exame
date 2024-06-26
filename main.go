package main

import (
	"github.com/MigFerro/exame/cookies"
	"github.com/MigFerro/exame/db"
	"github.com/MigFerro/exame/handlers"
	"github.com/MigFerro/exame/middleware"
	"github.com/MigFerro/exame/services"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {

	sqlxDB := db.ConnectToDB()
	authCookieStore := cookies.InitCookieStore()

	app := echo.New()

	app.Use(session.Middleware(authCookieStore))
	app.Use(middleware.WithAuthenticatedUser)
	app.Use(echomiddleware.Logger())

	app.Static("/static", "templates/static")

	exerciseService := services.ExerciseService{DB: sqlxDB}
	userService := services.UserService{DB: sqlxDB}

	homeHandler := handlers.HomeHandler{ExerciseService: &exerciseService}
	exerciseHandler := handlers.ExerciseHandler{ExerciseService: &exerciseService, UsersService: &userService}
	authHandler := handlers.AuthHandler{UserService: &userService}
	userHandler := handlers.UserHandler{UsersService: &userService}

	app.GET("/", homeHandler.HomeShow)
	app.GET("/year-category-list", homeHandler.YearExerciseCategoryListShow)

	// auth
	app.GET("/auth/:provider", authHandler.GetAuthProvider)
	app.GET("/auth/:provider/callback", authHandler.Login)
	app.GET("/auth/logout", authHandler.Logout)

	// users
	app.GET("/user/preplevel", userHandler.GetLoggedUserPrepLevel)
	app.GET("/user/points", userHandler.GetLoggedUserPoints)

	//exercises
	app.GET("/exercise", exerciseHandler.ShowExerciseToSolve)
	app.GET("/exercises", exerciseHandler.ShowExerciseList)
	app.GET("/exercises/create", exerciseHandler.ShowExerciseUpsert)
	app.GET("/exercises/:id/update", exerciseHandler.ShowExerciseUpsert)
	app.GET("/exercises/:id", exerciseHandler.ShowExerciseDetail)
	app.GET("/exercises/:id/result", exerciseHandler.ShowExerciseResult)
	app.GET("/exercises/:id/choices", exerciseHandler.ShowExerciseChoices)
	app.GET("/exercises/categories", exerciseHandler.ShowExerciseCategoriesList)
	app.GET("/exercises/category/:id", exerciseHandler.ShowExerciseCategoryDetail)
	app.GET("/exercises/category/create", exerciseHandler.ShowCreateExerciseCategory)
	app.GET("/exercises/category/:id/edit", exerciseHandler.ShowUpdateExerciseCategory)

	app.POST("/exercises/create", exerciseHandler.HandleExerciseUpsertJourney)
	app.POST("/exercises/:id/update", exerciseHandler.HandleExerciseUpsertJourney)
	app.POST("/exercises/:id", exerciseHandler.HandleExerciseSolve)
	app.POST("/exercises/category/create", exerciseHandler.CreateExerciseCategory)
	app.POST("/exercises/category/:id", exerciseHandler.UpdateExerciseCategory)

	app.Logger.Debug(app.Start(":3000"))
}
