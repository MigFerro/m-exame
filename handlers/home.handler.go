package handlers

import (
	"github.com/MigFerro/exame/services"
	errorsview "github.com/MigFerro/exame/templates/errors"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	ExerciseService *services.ExerciseService
}

func (h *HomeHandler) HomeShow(c echo.Context) error {
	id, err := h.ExerciseService.GetRandomExerciseId()
	exercise, err := h.ExerciseService.GetExerciseWithChoices(id)

	if err != nil {
		return render(c, errorsview.GeneralErrorPage())
	}

	return render(c, homeview.Show(exercise))
}

func (h *HomeHandler) YearExerciseCategoryListShow(c echo.Context) error {
	showList := c.QueryParam("show")
	showYear := c.QueryParam("showYear")

	categories, err := h.ExerciseService.GetCategoriesByYear(showYear)

	if err != nil {
		return render(c, errorsview.GeneralErrorPage())
	}

	return render(c, homeview.YearExerciseCategoryList(showList == "show", showYear, categories))
}
