package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"time"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/entities"
	"github.com/MigFerro/exame/services"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	ExerciseService *services.ExerciseService
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
		return h.saveExercise(c)
	}

	return nil
}

func (h *ExerciseHandler) ShowExerciseCreate(c echo.Context) error {

	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		return err
	}

	form := data.ExerciseCreationForm{
		Choices: []data.ExerciseChoice{
			{Value: "", IsSolution: true}, // default first choice is solution
			{Value: "", IsSolution: false},
			{Value: "", IsSolution: false},
			{Value: "", IsSolution: false},
		},
	}

	return render(c, exerciseview.ShowCreate(form, categories))
}

func (h *ExerciseHandler) ShowExerciseList(c echo.Context) error {

	var exercises []entities.ExerciseEntity
	err := h.ExerciseService.DB.Select(&exercises, `SELECT * FROM exercises`)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowIndex(exercises))
}

func (h *ExerciseHandler) ShowExerciseHomepage(c echo.Context) error {
	exerciseId := c.Param("id")

	exercise := h.ExerciseService.GetExerciseWithChoices(exerciseId)

	return render(c, homeview.HomepageExercise(exercise))
}

func (h *ExerciseHandler) ShowExerciseDetail(c echo.Context) error {
	exerciseId := c.Param("id")

	exercise := h.ExerciseService.GetExerciseWithChoices(exerciseId)

	return render(c, exerciseview.ShowDetail(exercise))
}

func (h *ExerciseHandler) ShowExerciseChoices(c echo.Context) error {
	exerciseId := c.Param("id")

	exerciseChoices := []entities.ExerciseChoiceEntity{}

	err := h.ExerciseService.DB.Select(&exerciseChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise choices from database: ", err)
		return err
	}

	choices := data.ExerciseChoices{
		Choices:    exerciseChoices,
		ExerciseId: exerciseId,
	}

	return render(c, exerciseview.ShowExerciseChoices(choices))
}

func (h *ExerciseHandler) HandleExerciseSolve(c echo.Context) error {
	exerciseId := c.Param("id")
	formChoice := c.Request().FormValue("choice")
	at := c.Request().FormValue("at")

	loggedUser, ok := c.Request().Context().Value("authUser").(*data.LoggedUser)

	exercise := entities.ExerciseEntity{}
	var isSolution bool

	tx := h.ExerciseService.DB.MustBegin()
	_ = tx.Get(&exercise, "SELECT * FROM exercises WHERE id = $1", exerciseId)
	_ = tx.Get(&isSolution, "SELECT is_solution FROM exercise_choices WHERE id = $1", formChoice)

	solvedData := data.ExerciseSolved{
		ExerciseId: exerciseId,
		IsSolution: isSolution,
		NextId:     h.ExerciseService.GetRandomExerciseId(),
		At:         at,
	}

	if !ok {
		return render(c, exerciseview.SolvedResult(solvedData))
	}

	exerciseUser := entities.ExerciseUserEntity{}
	_ = tx.Get(&exerciseUser, "SELECT * FROM exercise_users WHERE exercise_id = $1 AND user_id = $2", exerciseId, loggedUser.Id)

	now := time.Now()

	if exerciseUser == (entities.ExerciseUserEntity{}) {
		if isSolution {
			tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, last_attempted_at, first_solved_at, last_solved_at) VALUES ($1, $2, $3, $4, $5)", loggedUser.Id, exerciseId, now, now, now)
		} else {
			tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, last_attempted_at) VALUES ($1, $2, $3)", loggedUser.Id, exerciseId, now)
		}
	} else {
		if isSolution {
			if exerciseUser.FirstSolvedAt.Valid {
				tx.MustExec("UPDATE exercise_users SET (last_attempted_at, last_solved_at) = ($1, $2) WHERE user_id = $3 AND exercise_id = $4", now, now, loggedUser.Id, exerciseId)
			} else {
				tx.MustExec("UPDATE exercise_users SET (last_attempted_at, first_solved_at, last_solved_at) = ($1, $2, $3) WHERE user_id = $4 AND exercise_id = $5", now, now, now, loggedUser.Id, exerciseId)
			}
		} else {
			tx.MustExec("UPDATE exercise_users SET last_attempted_at = $1 WHERE user_id = $2 AND exercise_id = $3", now, loggedUser.Id, exerciseId)
		}
	}
	tx.Commit()

	return render(c, exerciseview.SolvedResult(solvedData))
}

func (h *ExerciseHandler) ShowExerciseCategoriesList(c echo.Context) error {
	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		fmt.Println("Error retrieving exercise categories from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowCategoriesIndex(categories))
}

func (h *ExerciseHandler) getExerciseCreationForm(c echo.Context) (data.ExerciseCreationForm, error) {

	authUser, ok := getAuthenticatedUser(c.Request().Context())

	if !ok {
		return data.ExerciseCreationForm{}, errors.New("No user authenticated. Can't create exercise.")
	}

	formPreviewText := c.Request().FormValue("problem_text")
	choices := []data.ExerciseChoice{}
	sol, _ := strconv.Atoi(c.Request().FormValue("choice_solution"))
	isSol := false
	for i := 0; i < 4; i++ {
		value := c.Request().FormValue("choice" + strconv.Itoa(i))
		if i == sol {
			isSol = true
		} else {
			isSol = false
		}
		choices = append(choices, data.ExerciseChoice{
			Value:      value,
			IsSolution: isSol,
		})
	}

	categoryIid, _ := strconv.Atoi(c.Request().FormValue("category"))
	var category string
	res := h.ExerciseService.DB.QueryRow("SELECT category from exercise_categories WHERE iid = $1", categoryIid)
	res.Scan(&category)
	exameYear := c.Request().FormValue("exame_year")
	exameFase := c.Request().FormValue("exame_fase")

	formResponse := data.ExerciseCreationForm{
		ProblemText: formPreviewText,
		Choices:     choices,
		Category: data.ExerciseCategory{
			Iid:      categoryIid,
			Category: category,
		},
		ExameYear: exameYear,
		ExameFase: exameFase,
		CreatedBy: authUser.Id,
	}

	return formResponse, nil

}

func (h *ExerciseHandler) exercisePreviewShow(c echo.Context) error {
	exerciseForm, err := h.getExerciseCreationForm(c)

	if err != nil {
		return err
	}

	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowCreate(exerciseForm, categories))
}

func (h *ExerciseHandler) exerciseSaveConfirmationPreviewShow(c echo.Context) error {
	exerciseForm, err := h.getExerciseCreationForm(c)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowSaveConfirmationPreview(exerciseForm))
}

func (h *ExerciseHandler) saveExercise(c echo.Context) error {
	exerciseForm, err := h.getExerciseCreationForm(c)

	if err != nil {
		return err
	}

	err = h.ExerciseService.SaveExercise(exerciseForm)

	return render(c, exerciseview.ExerciseSavedSuccessShow())
}
