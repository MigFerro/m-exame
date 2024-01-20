package entities

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ExerciseEntity struct {
	Id          uuid.UUID     `db:"id"`
	ProblemText string        `db:"problem_text"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   sql.NullTime  `db:"updated_at"`
	CreatedBy   uuid.UUID     `db:"created_by"`
	UpdatedBy   uuid.NullUUID `db:"updated_by"`
}

type ExerciseChoiceEntity struct {
	Id         uuid.UUID     `db:"id"`
	ExerciseId uuid.UUID     `db:"exercise_id"`
	Value      string        `db:"value"`
	IsSolution bool          `db:"solution"`
	CreatedAt  time.Time     `db:"created_at"`
	UpdatedAt  sql.NullTime  `db:"updated_at"`
	CreatedBy  uuid.UUID     `db:"created_by"`
	UpdatedBy  uuid.NullUUID `db:"updated_by"`
}

type ExerciseUserEntity struct {
	UserId     uuid.UUID    `db:"user_id"`
	ExerciseId uuid.UUID    `db:"exercise_id"`
	Solved     bool         `db:"solved"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  sql.NullTime `db:"updated_at"`
}
