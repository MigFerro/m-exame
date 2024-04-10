package entities

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ExerciseEntity struct {
	Id          uuid.UUID     `db:"id"`
	ProblemText string        `db:"problem_text"`
	CategoryIid int           `db:"category_iid"`
	ExameYear   string        `db:"exame"`
	ExameFase   string        `db:"fase"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   sql.NullTime  `db:"updated_at"`
	CreatedBy   uuid.UUID     `db:"created_by"`
	UpdatedBy   uuid.NullUUID `db:"updated_by"`
}

type ExerciseChoiceEntity struct {
	Id         uuid.UUID     `db:"id"`
	ExerciseId uuid.UUID     `db:"exercise_id"`
	Value      string        `db:"value"`
	IsSolution bool          `db:"is_solution"`
	CreatedAt  time.Time     `db:"created_at"`
	UpdatedAt  sql.NullTime  `db:"updated_at"`
	CreatedBy  uuid.UUID     `db:"created_by"`
	UpdatedBy  uuid.NullUUID `db:"updated_by"`
}

type ExerciseWithChoicesEntity struct {
	ExerciseEntity
	Category ExerciseCategoryEntity `db:"category"`
	Choices  []ExerciseChoiceEntity
}

type ExerciseCategoryEntity struct {
	Iid       int          `db:"iid"`
	Category  string       `db:"category"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	Year      string       `db:"year"`
}

type ExerciseUserEntity struct {
	UserId           uuid.UUID    `db:"user_id"`
	ExerciseId       uuid.UUID    `db:"exercise_id"`
	FirstAttemptedAt time.Time    `db:"first_attempted_at"`
	LastAttemptedAt  sql.NullTime `db:"last_attempted_at"`
	FirstSolvedAt    sql.NullTime `db:"first_solved_at"`
	LastSolvedAt     sql.NullTime `db:"last_solved_at"`
}
