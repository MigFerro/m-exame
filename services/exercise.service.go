package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MigFerro/exame/data"
	"github.com/MigFerro/exame/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ExerciseService struct {
	DB *sqlx.DB
}

func (s *ExerciseService) GetExerciseList() (uuid.UUIDs, error) {

	var exerciseIds uuid.UUIDs
	err := s.DB.Select(&exerciseIds, `SELECT id FROM exercises`)

	return exerciseIds, err
}

func (s *ExerciseService) GetCategoriesByYear(year string) ([]data.ExerciseCategory, error) {

	var categories []data.ExerciseCategory
	err := s.DB.Select(&categories, `
		SELECT iid, category FROM exercise_categories ec
		WHERE ec.year = $1`, year)

	return categories, err
}

func (s *ExerciseService) GetExerciseListOfCategory(categoryIid string) (uuid.UUIDs, error) {

	var exerciseIds uuid.UUIDs
	err := s.DB.Select(&exerciseIds, `
		SELECT id FROM exercises e
		INNER JOIN exercise_categories ec
		ON ec.iid = e.category_iid
		AND ec.iid = $1
		`, categoryIid)

	return exerciseIds, err
}

func (s *ExerciseService) GetExerciseHistory(userId uuid.UUID) (uuid.UUIDs, error) {

	var exerciseIds uuid.UUIDs
	err := s.DB.Select(&exerciseIds,
		`
		SELECT exercise_id FROM exercise_users
		WHERE user_id = $1
		ORDER BY last_attempted_at DESC
	`, userId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return exerciseIds, err
	}

	return exerciseIds, err
}

func (s *ExerciseService) SolveExercise(userId uuid.UUID, exerciseId string, choiceId string) (data.ExerciseSolved, error) {
	var exercise entities.ExerciseWithChoicesEntity
	var solutionId uuid.UUID

	tx := s.DB.MustBegin()
	err := tx.Get(&exercise, `
        SELECT
            e.*,
            cat.iid AS category_iid,
            cat.category AS "category.category",
            cat.created_at AS "category.created_at",
            cat.updated_at AS "category.updated_at"
        FROM
            exercises e
        LEFT JOIN
            exercise_categories cat ON e.category_iid = cat.iid
		WHERE e.id = $1`, exerciseId)

	dbChoices := []entities.ExerciseChoiceEntity{}
	err = s.DB.Select(&dbChoices, `
	SELECT * FROM exercise_choices
	WHERE exercise_id = $1
	ORDER BY random()`, exercise.Id)

	exercise.Choices = dbChoices

	err = tx.Get(&solutionId, "SELECT id FROM exercise_choices WHERE exercise_id = $1 AND is_solution = TRUE", exerciseId)

	exerciseUser := entities.ExerciseUserEntity{}
	err = tx.Get(&exerciseUser, "SELECT * FROM exercise_users WHERE exercise_id = $1 AND user_id = $2", exerciseId, userId)

	now := time.Now()

	isSolution := solutionId.String() == choiceId

	repeated := exerciseUser != (entities.ExerciseUserEntity{})

	points := 0
	if isSolution {
		points = 10
		if repeated {
			points = 5
		}
	} else {
		points = -2
	}

	if exerciseUser == (entities.ExerciseUserEntity{}) {
		if isSolution {
			tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, last_attempted_at, first_solved_at, last_solved_at, choice_selected, points_gained) VALUES ($1, $2, $3, $4, $5, $6, $7)", userId, exerciseId, now, now, now, choiceId, points)
		} else {
			tx.MustExec("INSERT INTO exercise_users (user_id, exercise_id, last_attempted_at, choice_selected, points_gained) VALUES ($1, $2, $3, $4, $5)", userId, exerciseId, now, choiceId, points)
		}
	} else {
		if isSolution {
			if exerciseUser.FirstSolvedAt.Valid {
				tx.MustExec("UPDATE exercise_users SET (last_attempted_at, last_solved_at, choice_selected, points_gained) = ($1, $2, $3, $4) WHERE user_id = $5 AND exercise_id = $6", now, now, choiceId, points, userId, exerciseId)
			} else {
				tx.MustExec("UPDATE exercise_users SET (last_attempted_at, first_solved_at, last_solved_at, choice_selected, points_gained) = ($1, $2, $3, $4, $5) WHERE user_id = $6 AND exercise_id = $7", now, now, now, choiceId, points, userId, exerciseId)
			}
		} else {
			tx.MustExec("UPDATE exercise_users SET (last_attempted_at, choice_selected, points_gained) = ($1, $2, $3) WHERE user_id = $4 AND exercise_id = $5", now, choiceId, points, userId, exerciseId)
		}
	}
	tx.Commit()

	s.updateUserPoints(userId, points)

	solvedData := data.ExerciseSolved{
		Exercise:         exercise,
		ChoiceSelectedId: choiceId,
		ChoiceCorrectId:  solutionId.String(),
		IsSolution:       isSolution,
		Repeated:         repeated,
		Points:           points,
	}

	return solvedData, err
}

func (s *ExerciseService) updateUserPoints(userId uuid.UUID, points int) {
	if points == 0 {
		return
	}

	var currPoints int
	query := `SELECT points FROM user_points WHERE user_id = $1`
	err := s.DB.Get(&currPoints, query, userId)

	if err != nil {
		fmt.Println("here")
		fmt.Println(err)
		return
	}

	// upsert
	now := time.Now()
	query = `UPDATE user_points SET points=$1, updated_at=$2 WHERE user_id=$3`
	_, err = s.DB.Exec(query, currPoints+points, now, userId)

	if err != nil {
		fmt.Println(err)
	}
}

func (s *ExerciseService) GetExerciseResult(userId uuid.UUID, exerciseId string) (data.ExerciseSolved, error) {
	var exerciseUser entities.ExerciseUserEntity

	query := `
		SELECT * FROM exercise_users
		WHERE user_id = $1
		AND exercise_id = $2
	`

	err := s.DB.Get(&exerciseUser, query, userId, exerciseId)

	exerciseWithChoices, err := s.GetExerciseWithChoices(exerciseId)
	var correctChoiceId string
	for _, choice := range exerciseWithChoices.Choices {
		if choice.IsSolution {
			correctChoiceId = choice.Id.String()
		}
	}

	solvedData := data.ExerciseSolved{
		Exercise:         exerciseWithChoices,
		ChoiceSelectedId: exerciseUser.ChoiceSelected.String(),
		ChoiceCorrectId:  correctChoiceId,
		IsSolution:       exerciseUser.ChoiceSelected.String() == correctChoiceId,
		Points:           int(exerciseUser.PointsGained.Int32),
	}

	return solvedData, err
}

func (s *ExerciseService) GetExerciseWithChoices(exerciseId string) (entities.ExerciseWithChoicesEntity, error) {

	var ex entities.ExerciseWithChoicesEntity
	query := `
        SELECT
            e.*,
            cat.iid AS category_iid,
            cat.category AS "category.category",
            cat.created_at AS "category.created_at",
            cat.updated_at AS "category.updated_at"
        FROM
            exercises e
        LEFT JOIN
            exercise_categories cat ON e.category_iid = cat.iid
        WHERE
            e.id = $1
    `
	err := s.DB.Get(&ex, query, exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return entities.ExerciseWithChoicesEntity{}, err
	}

	exerciseChoices := []entities.ExerciseChoiceEntity{}

	err = s.DB.Select(&exerciseChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1
		ORDER BY random()`, exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return entities.ExerciseWithChoicesEntity{}, err
	}

	ex.Choices = exerciseChoices

	return ex, nil
}

func (s *ExerciseService) GetExerciseUpsertForm(exerciseId string) (*data.ExerciseUpsertForm, error) {
	exerciseWithChoices, err := s.GetExerciseWithChoices(exerciseId)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
		return &data.ExerciseUpsertForm{}, err
	}

	catRow := s.DB.QueryRowx(
		`SELECT * FROM exercise_categories
		WHERE iid = $1`, exerciseWithChoices.Category.Iid)

	var cat entities.ExerciseCategoryEntity
	err = catRow.StructScan(&cat)

	var choices []entities.ExerciseChoiceEntity
	err = s.DB.Select(&choices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exerciseWithChoices.Id)

	formChoices := []data.ExerciseChoice{}
	var c data.ExerciseChoice
	for _, choice := range choices {
		c = data.ExerciseChoice{
			Id:         choice.Id,
			Value:      choice.Value,
			IsSolution: choice.IsSolution,
		}
		formChoices = append(formChoices, c)
	}

	form := data.ExerciseUpsertForm{
		Id:          exerciseWithChoices.Id.String(),
		ProblemText: exerciseWithChoices.ProblemText,
		ExameYear:   exerciseWithChoices.ExameYear,
		ExameFase:   exerciseWithChoices.ExameFase,
		Category: data.ExerciseCategory{
			Iid:      cat.Iid,
			Category: cat.Category,
		},
		Choices: formChoices,
	}

	return &form, nil
}

// check if this is used
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

func (s *ExerciseService) UpdateExercise(exerciseForm *data.ExerciseUpsertForm) error {
	if exerciseForm.Id == "" {
		return errors.New("no valid exercise id")
	}

	now := time.Now()

	// Begin transaction
	tx := s.DB.MustBegin()

	// Save exercise
	_, err := tx.Exec(`UPDATE exercises
		SET (problem_text, category_iid, exame, fase, updated_by, updated_at) = ($1, $2, $3, $4, $5, $6)
		WHERE id = $7`, exerciseForm.ProblemText, exerciseForm.Category.Iid, exerciseForm.ExameYear, exerciseForm.ExameFase, exerciseForm.UpdatedBy, now, exerciseForm.Id)

	if err != nil {
		fmt.Println(err)
		return err
	}

	// Save exercise choices
	dbChoices := []entities.ExerciseChoiceEntity{}

	err = s.DB.Select(&dbChoices,
		`SELECT * FROM exercise_choices
		WHERE exercise_id = $1`, exerciseForm.Id)

	if err != nil {
		fmt.Println("Error retrieving exercise from database: ", err)
	}

	for i, choice := range exerciseForm.Choices {
		_, err = tx.Exec(`UPDATE exercise_choices SET
			(value, is_solution, updated_by, updated_at) = ($1, $2, $3, $4)
			WHERE id = $5
		`, choice.Value, choice.IsSolution, exerciseForm.UpdatedBy, now, dbChoices[i].Id)

		if err != nil {
			fmt.Println(err)
			return err
		}
	}

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

func (s *ExerciseService) GetExerciseCategory(iid string) (entities.ExerciseCategoryEntity, error) {
	catRow := s.DB.QueryRowx(
		`SELECT * FROM exercise_categories
		WHERE iid = $1`, iid)

	var category entities.ExerciseCategoryEntity
	err := catRow.StructScan(&category)

	if err != nil {
		fmt.Println("Error retrieving category from database: ", err)
		return entities.ExerciseCategoryEntity{}, err
	}

	return category, nil
}

func (s *ExerciseService) UpdateCategory(category entities.ExerciseCategoryEntity) error {
	// Begin transaction
	tx := s.DB.MustBegin()

	_, err := tx.Exec(`UPDATE exercise_categories
		SET (category, year) = ($1, $2)
		WHERE iid = $3`, category.Category, category.Year, category.Iid)

	// End transaction
	tx.Commit()

	return err
}

func (s *ExerciseService) SaveCategory(category string, year string) error {
	// Begin transaction
	tx := s.DB.MustBegin()
	tx.QueryRow("INSERT INTO exercise_categories (category, year) VALUES ($1, $2)", category, year)

	// End transaction
	tx.Commit()

	return nil
}
