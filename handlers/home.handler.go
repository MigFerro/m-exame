package handlers

import (
	"github.com/MigFerro/exame/services"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	ExerciseService *services.ExerciseService
}

func (h *HomeHandler) HomeShow(c echo.Context) error {
	return render(c, homeview.Show())
}

func (h *HomeHandler) ExameExerciseListShow(c echo.Context) error {
	showList := c.QueryParam("show")

	return render(c, homeview.ExameExerciseChoiceList(showList == "show"))
}
