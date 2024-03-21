package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"time"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/entities"
	"github.com/MigFerro/exame/services"
	errorsview "github.com/MigFerro/exame/templates/errors"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type ExerciseHandler struct {
	ExerciseService *services.ExerciseService
	UsersService    *services.UserService
}

func (h *ExerciseHandler) HandleExerciseUpsertJourney(c echo.Context) error {
	formAction := c.Request().FormValue("action")
	exerciseId := c.Param("id")

	if formAction == "preview" {
		return h.exercisePreviewShow(c)
	}

	if formAction == "save" {
		return h.exerciseSaveConfirmationPreviewShow(c)
	}

	if formAction == "back" {
		if exerciseId != "" {
			return h.ShowExerciseUpdateBack(c)
		}
		return h.ShowExerciseCreate(c)
	}

	if formAction == "confirm" {
		return h.saveExercise(c)
	}

	return nil
}

func (h *ExerciseHandler) ShowExerciseCreate(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	exerciseForm, err := h.getExerciseUpsertForm(c)
	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowCreate(exerciseForm, categories))
}

func (h *ExerciseHandler) ShowExerciseUpdate(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	exerciseId := c.Param("id")

	exercise, err := h.ExerciseService.GetExerciseUpsertForm(exerciseId)
	if err != nil {
		return err
	}

	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		return err
	}

	exercise.UpdatedBy = authUser.Id

	return render(c, exerciseview.ShowUpdate(exercise, categories))
}

func (h *ExerciseHandler) ShowExerciseUpdateBack(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	exerciseId := c.Param("id")
	exerciseForm, _ := h.getExerciseUpsertForm(c)
	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		return err
	}

	exerciseForm.Id = exerciseId

	return render(c, exerciseview.ShowUpdate(exerciseForm, categories))
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

	exercise, _ := h.ExerciseService.GetExerciseWithChoices(exerciseId)

	return render(c, homeview.HomepageExercise(exercise))
}

func (h *ExerciseHandler) ShowExerciseDetail(c echo.Context) error {
	exerciseId := c.Param("id")

	exercise, _ := h.ExerciseService.GetExerciseWithChoices(exerciseId)

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

func (h *ExerciseHandler) getExerciseUpsertForm(c echo.Context) (*data.ExerciseUpsertForm, error) {
	authUser, ok := getAuthenticatedUser(c.Request().Context())

	if !ok {
		return &data.ExerciseUpsertForm{}, errors.New("No user authenticated.")
	}

	formPreviewText := c.Request().FormValue("problem_text")
	choices := []data.ExerciseChoice{}
	sol, _ := strconv.Atoi(c.Request().FormValue("choice_solution"))
	for i := 0; i < 4; i++ {
		value := c.Request().FormValue("choice" + strconv.Itoa(i))
		choices = append(choices, data.ExerciseChoice{
			Value:      value,
			IsSolution: i == sol,
		})
	}

	categoryIid, _ := strconv.Atoi(c.Request().FormValue("category"))
	var category string
	res := h.ExerciseService.DB.QueryRow("SELECT category from exercise_categories WHERE iid = $1", categoryIid)
	res.Scan(&category)
	exameYear := c.Request().FormValue("exame_year")
	exameFase := c.Request().FormValue("exame_fase")

	formResponse := data.ExerciseUpsertForm{
		ProblemText: formPreviewText,
		Choices:     choices,
		Category: data.ExerciseCategory{
			Iid:      categoryIid,
			Category: category,
		},
		ExameYear: exameYear,
		ExameFase: exameFase,
		CreatedBy: authUser.Id,
		UpdatedBy: authUser.Id,
	}

	exerciseId := c.Param("id")
	if exerciseId != "" {
		formResponse.Id = exerciseId
	}

	return &formResponse, nil

}

func (h *ExerciseHandler) exercisePreviewShow(c echo.Context) error {
	exerciseForm, err := h.getExerciseUpsertForm(c)

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
	exerciseForm, err := h.getExerciseUpsertForm(c)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowSaveConfirmationPreview(exerciseForm))
}

func (h *ExerciseHandler) saveExercise(c echo.Context) error {
	exerciseForm, err := h.getExerciseUpsertForm(c)
	exerciseId := c.Param("id")

	if err != nil {
		return err
	}

	if exerciseId != "" {
		err = h.ExerciseService.UpdateExercise(exerciseForm)
	} else {
		err = h.ExerciseService.SaveExercise(exerciseForm)
	}

	if err != nil {
		return err
	}

	return render(c, exerciseview.ExerciseSavedSuccessShow())
}
