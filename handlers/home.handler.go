package handlers

import (
	"fmt"

	"github.com/MigFerro/exame/entities"
	homeview "github.com/MigFerro/exame/templates/home"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	DB *sqlx.DB
}

func (h *HomeHandler) HomeShow(c echo.Context) error {
	exercise, exerciseChoices := h.getHomepageExercise()
	return render(c, homeview.Show(exercise, exerciseChoices))
}

func (h *HomeHandler) getHomepageExercise() (entities.ExerciseEntity, []entities.ExerciseChoiceEntity) {

	exerciseRow := h.DB.QueryRowx(
		`SELECT * FROM exercises
		ORDER BY random()
		LIMIT 1
	`)

	var exercise entities.ExerciseEntity
	err := exerciseRow.StructScan(&exercise)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		// return err
	}

	exerciseChoices := []entities.ExerciseChoiceEntity{}

	err = h.DB.Select(&exerciseChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exercise.Id)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		// return err
	}

	return exercise, exerciseChoices
}
