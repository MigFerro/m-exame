package handlers

import (
	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/services"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	ExerciseService *services.ExerciseService
}

func (h *HomeHandler) HomeShow(c echo.Context) error {
	exerciseId := c.Param("id")
	if exerciseId == "" {
		exerciseId = h.ExerciseService.GetRandomExerciseId()
	}

	exercise := h.ExerciseService.GetExerciseWithChoices(exerciseId)

	exercises := data.ExerciseWithChoices{
		Exercise: exercise.Exercise,
		Choices:  exercise.Choices,
	}

	return render(c, homeview.Show(exercises))
}
