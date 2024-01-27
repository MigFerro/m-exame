package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MigFerro/exame/entities"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	"github.com/google/uuid"
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
		return handleExerciseSavePreview(c)
	}

	if formAction == "Confirmar" {
		return handleSaveExercise(c, h.DB)
	}

	return nil
}

func (h *ExerciseHandler) ExerciseCreateShow(c echo.Context) error {

	return render(c, exerciseview.ShowCreate("", []string{"", "", "", ""}))
}

func getExerciseForm(c echo.Context) (string, []string) {
	formPreviewText := c.Request().FormValue("problem_text")
	choices := []string{"", "", "", ""}
	for i := 0; i < 4; i++ {
		choice := c.Request().FormValue("choice" + strconv.Itoa(i))
		choices[i] = choice
	}

	return formPreviewText, choices

}

func exercisePreviewShow(c echo.Context) error {
	formPreviewText, choices := getExerciseForm(c)

	return render(c, exerciseview.ShowCreate(formPreviewText, choices))
}

func handleExerciseSavePreview(c echo.Context) error {
	formPreviewText, choices := getExerciseForm(c)

	return render(c, exerciseview.ShowSavePreview(formPreviewText, choices))
}

func saveExerciseChoices(exerciseId uuid.UUID, choices []string, authUserId uuid.UUID, tx *sqlx.Tx) {
	isSolution := false
	numOfChoices := 4

	valueStrings := make([]string, 0, numOfChoices)
	valueArgs := make([]interface{}, 0, numOfChoices*4)

	for i, choice := range choices {
		if i == 0 {
			isSolution = true
		} else {
			isSolution = false
		}

		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", 4*i+1, 4*i+2, 4*i+3, 4*i+4))
		valueArgs = append(valueArgs, choice)
		valueArgs = append(valueArgs, isSolution)
		valueArgs = append(valueArgs, authUserId)
		valueArgs = append(valueArgs, exerciseId)
	}

	query := fmt.Sprintf("INSERT INTO exercise_choices (value, solution, created_by, exercise_id) VALUES %s", strings.Join(valueStrings, ","))
	fmt.Println(query)

	tx.MustExec(query, valueArgs...)
}

func handleSaveExercise(c echo.Context, db *sqlx.DB) error {
	formPreviewText, choices := getExerciseForm(c)

	authUser := getAuthenticatedUser(c.Request().Context())

	exercise := entities.ExerciseEntity{
		ProblemText: formPreviewText,
		CreatedBy:   authUser.Id,
	}

	tx := db.MustBegin()
	res := tx.QueryRow("INSERT INTO exercises (problem_text, created_by) VALUES ($1, $2) RETURNING id", exercise.ProblemText, exercise.CreatedBy)
	var exerciseId uuid.UUID
	res.Scan(&exerciseId)
	saveExerciseChoices(exerciseId, choices, authUser.Id, tx)
	tx.Commit()

	return render(c, exerciseview.ShowExerciseSaved())
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
	var isSolution bool

	tx := h.DB.MustBegin()
	_ = tx.Get(&exerciseUser, "SELECT * FROM exercise_users WHERE exercise_id = $1 AND user_id = $2", exerciseId, authUser.Id)
	_ = tx.Get(&exercise, "SELECT * FROM exercises WHERE id = $1", exerciseId)
	_ = tx.Get(&isSolution, "SELECT solution FROM exercise_choices WHERE id = $1", formChoice)

	solved := exerciseUser.Solved || isSolution

	if exerciseUser == (entities.ExerciseUserEntity{}) {
		tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, solved, updated_at) VALUES ($1, $2, $3, $4)", authUser.Id, exerciseId, solved, time.Now())
	} else {
		tx.MustExec("UPDATE exercise_users SET (solved, updated_at) = ($1, $2) WHERE user_id = $3 AND exercise_id = $4", solved, time.Now(), authUser.Id, exerciseId)
	}
	tx.Commit()

	return render(c, exerciseview.Solve(isSolution))
}
