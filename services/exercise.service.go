package services

import (
	"fmt"
	"strings"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ExerciseService struct {
	DB *sqlx.DB
}

func (s *ExerciseService) GetExerciseWithChoices(exerciseId string) data.ExerciseWithChoices {
	exerciseRow := s.DB.QueryRowx(
		`SELECT * FROM exercises
		WHERE id = $1`, exerciseId)

	var exercise entities.ExerciseEntity
	err := exerciseRow.StructScan(&exercise)

	exerciseChoices := []entities.ExerciseChoiceEntity{}

	err = s.DB.Select(&exerciseChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1
		ORDER BY random()`, exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
	}

	exerciseWithChoices := data.ExerciseWithChoices{
		Choices:  exerciseChoices,
		Exercise: exercise,
	}

	return exerciseWithChoices
}

func (s *ExerciseService) GetExerciseUpsertForm(exerciseId string) *data.ExerciseUpsertForm {
	exerciseWithChoices := s.GetExerciseWithChoices(exerciseId)

	catRow := s.DB.QueryRowx(
		`SELECT * FROM exercise_categories
		WHERE iid = $1`, exerciseWithChoices.Exercise.CategoryIid)

	var cat entities.ExerciseCategoryEntity
	err := catRow.StructScan(&cat)

	var choices []entities.ExerciseChoiceEntity
	err = s.DB.Select(&choices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exerciseWithChoices.Exercise.Id)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
	}

	formChoices := []data.ExerciseChoice{}
	var c data.ExerciseChoice
	for _, choice := range choices {
		c = data.ExerciseChoice{
			Value:      choice.Value,
			IsSolution: choice.IsSolution,
		}
		formChoices = append(formChoices, c)
	}

	form := data.ExerciseUpsertForm{
		Id:          exerciseWithChoices.Exercise.Id.String(),
		ProblemText: exerciseWithChoices.Exercise.ProblemText,
		ExameYear:   exerciseWithChoices.Exercise.ExameYear,
		ExameFase:   exerciseWithChoices.Exercise.ExameFase,
		Category: data.ExerciseCategory{
			Iid:      cat.Iid,
			Category: cat.Category,
		},
		Choices: formChoices,
	}

	return &form
}

func (s *ExerciseService) GetPreviouslyAttemptedExercise(exerciseId string, userId string) string {

	exerciseUserRow := s.DB.QueryRowx(
		`SELECT exercise_id FROM exercise_users
		WHERE id = $1`, exerciseId)

	fmt.Println(exerciseUserRow)
	return ""

}

func (s *ExerciseService) GetRandomExerciseId() string {
	exerciseRow := s.DB.QueryRowx(
		`SELECT id FROM exercises
		ORDER BY random()
		LIMIT 1`)

	var exerciseId string
	_ = exerciseRow.Scan(&exerciseId)

	return exerciseId
}

func (s *ExerciseService) SaveExercise(exerciseForm *data.ExerciseUpsertForm) error {
	// Begin transaction
	tx := s.DB.MustBegin()
	res := tx.QueryRow("INSERT INTO exercises (problem_text, category_iid, exame, fase, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", exerciseForm.ProblemText, exerciseForm.Category.Iid, exerciseForm.ExameYear, exerciseForm.ExameFase, exerciseForm.CreatedBy)
	var exerciseId uuid.UUID
	res.Scan(&exerciseId)
	fmt.Println(exerciseId)

	// Save exercise choices
	valueStrings := make([]string, 0, len(exerciseForm.Choices))
	valueArgs := make([]interface{}, 0, len(exerciseForm.Choices)*4)

	for i, choice := range exerciseForm.Choices {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", 4*i+1, 4*i+2, 4*i+3, 4*i+4))
		valueArgs = append(valueArgs, choice.Value)
		valueArgs = append(valueArgs, choice.IsSolution)
		valueArgs = append(valueArgs, exerciseForm.CreatedBy)
		valueArgs = append(valueArgs, exerciseId)
	}

	query := fmt.Sprintf("INSERT INTO exercise_choices (value, is_solution, created_by, exercise_id) VALUES %s", strings.Join(valueStrings, ","))
	tx.MustExec(query, valueArgs...)

	// End transaction
	tx.Commit()

	return nil
}

func (s *ExerciseService) GetAllCategories() ([]entities.ExerciseCategoryEntity, error) {
	var categories []entities.ExerciseCategoryEntity
	err := s.DB.Select(&categories, `SELECT * FROM exercise_categories`)

	if err != nil {
		fmt.Println("Error retrieving exercise categories from database: ", err)
		return []entities.ExerciseCategoryEntity{}, err
	}

	return categories, nil
}
