package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MigFerro/exame/entities"
	"github.com/MigFerro/exame/shared"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	DB *sqlx.DB
}

func (h *ExerciseHandler) HandleExerciseCreateJourney(c echo.Context) error {
	formAction := c.Request().FormValue("action")

	if formAction == "Preview" {
		return h.exercisePreviewShow(c)
	}

	if formAction == "Save" {
		return h.exerciseSaveConfirmationPreviewShow(c)
	}

	if formAction == "Confirmar" {
		return h.saveExerciseWithChoices(c)
	}

	return nil
}

func (h *ExerciseHandler) ExerciseCreateShow(c echo.Context) error {

	err, categories := getAllCategories(h.DB)

	if err != nil {
		return err
	}

	form := shared.ExerciseFormResponse{
		Choices: []shared.ExerciseChoice{
			{Value: "", IsSolution: true}, // default first choice is solution
			{Value: "", IsSolution: false},
			{Value: "", IsSolution: false},
			{Value: "", IsSolution: false},
		},
	}

	return render(c, exerciseview.ShowCreate(form, categories))
}

func (h *ExerciseHandler) getExerciseForm(c echo.Context) shared.ExerciseFormResponse {
	formPreviewText := c.Request().FormValue("problem_text")
	choices := []shared.ExerciseChoice{}
	sol, _ := strconv.Atoi(c.Request().FormValue("choice_solution"))
	isSol := false
	for i := 0; i < 4; i++ {
		value := c.Request().FormValue("choice" + strconv.Itoa(i))
		if i == sol {
			isSol = true
		} else {
			isSol = false
		}
		choices = append(choices, shared.ExerciseChoice{
			Value:      value,
			IsSolution: isSol,
		})
	}

	categoryIid, _ := strconv.Atoi(c.Request().FormValue("category"))
	var category string
	res := h.DB.QueryRow("SELECT category from exercise_categories WHERE iid = $1", categoryIid)
	res.Scan(&category)
	exameYear := c.Request().FormValue("exame_year")
	exameFase := c.Request().FormValue("exame_fase")

	formResponse := shared.ExerciseFormResponse{
		ProblemText: formPreviewText,
		Choices:     choices,
		Category: shared.ExerciseCategory{
			Iid:      categoryIid,
			Category: category,
		},
		ExameYear: exameYear,
		ExameFase: exameFase,
	}

	return formResponse

}

func (h *ExerciseHandler) exercisePreviewShow(c echo.Context) error {
	exerciseForm := h.getExerciseForm(c)
	err, categories := getAllCategories(h.DB)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowCreate(exerciseForm, categories))
}

func (h *ExerciseHandler) exerciseSaveConfirmationPreviewShow(c echo.Context) error {
	exerciseForm := h.getExerciseForm(c)

	return render(c, exerciseview.ShowSaveConfirmationPreview(exerciseForm))
}

func saveExerciseChoices(exerciseId uuid.UUID, choices []shared.ExerciseChoice, authUserId uuid.UUID, tx *sqlx.Tx) {
	valueStrings := make([]string, 0, len(choices))
	valueArgs := make([]interface{}, 0, len(choices)*4)

	for i, choice := range choices {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", 4*i+1, 4*i+2, 4*i+3, 4*i+4))
		valueArgs = append(valueArgs, choice.Value)
		valueArgs = append(valueArgs, choice.IsSolution)
		valueArgs = append(valueArgs, authUserId)
		valueArgs = append(valueArgs, exerciseId)
	}

	query := fmt.Sprintf("INSERT INTO exercise_choices (value, is_solution, created_by, exercise_id) VALUES %s", strings.Join(valueStrings, ","))

	tx.MustExec(query, valueArgs...)
}

func (h *ExerciseHandler) saveExerciseWithChoices(c echo.Context) error {
	exerciseForm := h.getExerciseForm(c)

	authUser := getAuthenticatedUser(c.Request().Context())

	exercise := entities.ExerciseEntity{
		ProblemText: exerciseForm.ProblemText,
		CategoryIid: exerciseForm.Category.Iid,
		ExameYear:   exerciseForm.ExameYear,
		ExameFase:   exerciseForm.ExameFase,
		CreatedBy:   authUser.Id,
	}

	tx := h.DB.MustBegin()
	res := tx.QueryRow("INSERT INTO exercises (problem_text, category_iid, exame, fase, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", exercise.ProblemText, exercise.CategoryIid, exercise.ExameYear, exercise.ExameFase, exercise.CreatedBy)
	var exerciseId uuid.UUID
	res.Scan(&exerciseId)
	fmt.Println(exerciseId)
	saveExerciseChoices(exerciseId, exerciseForm.Choices, authUser.Id, tx)
	tx.Commit()

	return render(c, exerciseview.ExerciseSavedSuccessShow())
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
	_ = tx.Get(&isSolution, "SELECT is_solution FROM exercise_choices WHERE id = $1", formChoice)

	solved := exerciseUser.Solved || isSolution

	if exerciseUser == (entities.ExerciseUserEntity{}) {
		tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, solved, updated_at, times_attempted) VALUES ($1, $2, $3, $4, $5)", authUser.Id, exerciseId, solved, time.Now(), 1)
	} else {
		if !exerciseUser.Solved && solved {
			tx.MustExec("UPDATE exercise_users SET (solved, updated_at, times_attempted) = ($1, $2, $3) WHERE user_id = $4 AND exercise_id = $5", solved, time.Now(), authUser.Id, exerciseId, exerciseUser.TimesAttempted+1)
		}
	}
	tx.Commit()

	return render(c, exerciseview.Solve(isSolution))
}

func (h *ExerciseHandler) ExerciseCategoriesShow(c echo.Context) error {
	err, categories := getAllCategories(h.DB)

	if err != nil {
		fmt.Println("Error retrieving exercise categories from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowCategoriesIndex(categories))
}

func getAllCategories(db *sqlx.DB) (error, []entities.ExerciseCategoryEntity) {
	var categories []entities.ExerciseCategoryEntity
	err := db.Select(&categories, `SELECT * FROM exercise_categories`)

	if err != nil {
		fmt.Println("Error retrieving exercise categories from database: ", err)
		return err, []entities.ExerciseCategoryEntity{}
	}

	return nil, categories
}
