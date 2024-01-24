package handlers

import (
	"fmt"
	"time"

	"github.com/MigFerro/exame/entities"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	DB *sqlx.DB
}

func (h *ExerciseHandler) HandleExerciseCreate(c echo.Context) error {
	formAction := c.Request().FormValue("action")

	if formAction == "Preview" {
		return exercisePreviewShow(c)
	}

	if formAction == "Save" {
		return nil
	}

	return nil
}

func (h *ExerciseHandler) ExerciseCreateShow(c echo.Context) error {

	return render(c, exerciseview.ShowCreate(""))
}

func exercisePreviewShow(c echo.Context) error {
	formPreviewText := c.Request().FormValue("problem_text")

	return render(c, exerciseview.ShowCreate(formPreviewText))
}

func (h *ExerciseHandler) ExerciseListShow(c echo.Context) error {

	var exercises []entities.ExerciseEntity
	err := h.DB.Select(&exercises, `SELECT * FROM exercises`)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowIndex(exercises))
}

func (h *ExerciseHandler) ExerciseDetailShow(c echo.Context) error {
	exerciseId := c.Param("id")

	exerciseRow := h.DB.QueryRowx(
		`SELECT * FROM exercises
		WHERE id = $1`, exerciseId)

	var exercise entities.ExerciseEntity
	err := exerciseRow.StructScan(&exercise)

	exerciseChoices := []entities.ExerciseChoiceEntity{}

	err = h.DB.Select(&exerciseChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowDetail(exercise, exerciseChoices))
}

func (h *ExerciseHandler) ExerciseSolve(c echo.Context) error {
	exerciseId := c.Param("id")
	formChoice := c.Request().FormValue("choice")

	authUser, ok := c.Request().Context().Value("authUser").(*entities.AuthUser)
	if !ok {
		fmt.Println("Error retrieving auth user")
		return nil
	}

	exercise := entities.ExerciseEntity{}
	exerciseUser := entities.ExerciseUserEntity{}
	var solution string

	tx := h.DB.MustBegin()
	_ = tx.Get(&exerciseUser, "SELECT * FROM exercise_users WHERE exercise_id = $1 AND user_id = $2", exerciseId, authUser.Id)
	_ = tx.Get(&exercise, "SELECT * FROM exercises WHERE id = $1", exerciseId)
	_ = tx.Get(&solution, "SELECT value FROM exercise_choices WHERE exercise_id = $1 AND solution = true", exerciseId)

	correctAns := solution == formChoice
	solved := exerciseUser.Solved || correctAns

	if exerciseUser == (entities.ExerciseUserEntity{}) {
		tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, solved, updated_at) VALUES ($1, $2, $3, $4)", authUser.Id, exerciseId, solved, time.Now())
	} else {
		tx.MustExec("UPDATE exercise_users SET (solved, updated_at) = ($1, $2) WHERE user_id = $3 AND exercise_id = $4", solved, time.Now(), authUser.Id, exerciseId)
	}
	tx.Commit()

	return render(c, exerciseview.Solve(correctAns))
}
