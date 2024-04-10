package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/entities"
	"github.com/MigFerro/exame/services"
	errorsview "github.com/MigFerro/exame/templates/errors"
	exerciseview "github.com/MigFerro/exame/templates/exercise"
	"github.com/google/uuid"
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
	category := c.QueryParam("category")

	var exerciseIds uuid.UUIDs

	if category != "" {
		exerciseIds, _ = h.ExerciseService.GetExerciseListOfCategory(category)
	} else {
		exerciseIds, _ = h.ExerciseService.GetExerciseList()
	}

	return render(c, exerciseview.ShowIndex(exerciseIds))
}

func (h *ExerciseHandler) ShowTest(c echo.Context) error {
	testType := c.QueryParam("type")
	loggedUser, ok := c.Request().Context().Value("authUser").(*data.LoggedUser)

	if !ok {
		return errors.New("No logged user")
	}

	exercises, err := h.ExerciseService.GetTestExercises(testType, loggedUser.Id)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowTest(exercises))
}

func (h *ExerciseHandler) SolveTest(c echo.Context) error {
	loggedUser, ok := c.Request().Context().Value("authUser").(*data.LoggedUser)

	if !ok {
		return errors.New("No logged user")
	}

	answers := h.readTestForm(c)
	res, err := h.ExerciseService.EvaluateAndSaveTest(loggedUser.Id, answers)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowTestResult(res))
}

func (h *ExerciseHandler) readTestForm(c echo.Context) []data.ExerciseAnswer {
	exerciseCount, _ := strconv.Atoi(c.Request().FormValue("exercise_count"))

	answers := []data.ExerciseAnswer{}
	ans := data.ExerciseAnswer{}

	for i := 0; i < exerciseCount; i++ {
		id := c.Request().FormValue("exercise-" + strconv.Itoa(i))
		choiceId := c.Request().FormValue("choice-" + strconv.Itoa(i))

		ans.Id = id
		ans.ChoiceId = choiceId

		answers = append(answers, ans)
	}

	return answers

}

func (h *ExerciseHandler) ShowExerciseHistory(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	exerciseIds, err := h.ExerciseService.GetExerciseHistory(authUser.Id)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowHistory(exerciseIds))
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

	loggedUser, ok := c.Request().Context().Value("authUser").(*data.LoggedUser)

	if !ok {
		return errors.New("No user logged in")
	}

	result, err := h.ExerciseService.SolveExercise(loggedUser.Id, exerciseId, formChoice, true)

	if err != nil {
		fmt.Println("Error saving exercise in database: ", err)
		return err
	}

	return render(c, exerciseview.SolvedResult(result))
}

func (h *ExerciseHandler) ShowExerciseCategoriesList(c echo.Context) error {
	categories, err := h.ExerciseService.GetAllCategories()

	if err != nil {
		fmt.Println("Error retrieving exercise categories from database: ", err)
		return err
	}

	return render(c, exerciseview.ShowCategoriesIndex(categories))
}

func (h *ExerciseHandler) ShowExerciseCategoryDetail(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())
	userRole := h.UsersService.GetUserRole(authUser.Id)

	categoryIid := c.Param("id")

	category, err := h.ExerciseService.GetExerciseCategory(categoryIid)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowCategoryDetail(category, userRole == "admin"))
}

func (h *ExerciseHandler) ShowCreateExerciseCategory(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	return render(c, exerciseview.ShowCategoryCreate())
}

func (h *ExerciseHandler) ShowUpdateExerciseCategory(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	categoryIid := c.Param("id")

	category, err := h.ExerciseService.GetExerciseCategory(categoryIid)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowUpdateCategory(category))
}

func (h *ExerciseHandler) UpdateExerciseCategory(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	categoryIid, _ := strconv.Atoi(c.Param("id"))
	categoryName := c.Request().FormValue("category")
	year := c.Request().FormValue("year")

	category := entities.ExerciseCategoryEntity{
		Iid:      categoryIid,
		Category: categoryName,
		Year:     year,
	}

	err := h.ExerciseService.UpdateCategory(category)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return render(c, exerciseview.ShowUpdateCategorySuccess(category))
}

func (h *ExerciseHandler) CreateExerciseCategory(c echo.Context) error {
	authUser, _ := getAuthenticatedUser(c.Request().Context())

	userRole := h.UsersService.GetUserRole(authUser.Id)

	if userRole != "admin" {
		return render(c, errorsview.PermissionDenied())
	}

	category := c.Request().FormValue("category")
	year := c.Request().FormValue("category_year")

	err := h.ExerciseService.SaveCategory(category, year)

	if err != nil {
		return err
	}

	return render(c, exerciseview.ShowCategoryCreateSuccess(category, year))
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
