package data

import (
	"github.com/MigFerro/exame/entities"
	"github.com/google/uuid"
)

type LoggedUser struct {
	Id     uuid.UUID
	AuthId string
	Email  string
	Name   string
}

type ExerciseChoices struct {
	Choices    []entities.ExerciseChoiceEntity
	ExerciseId string
}

type ExerciseSolved struct {
	Exercise           entities.ExerciseWithChoicesEntity
	IsSolution         bool
	ChoiceSelectedId   string
	ChoiceCorrectId    string
	NextId             string
	Repeated           bool
	Points             int
	PreviousExerciseId uuid.NullUUID
	NextExerciseId     uuid.NullUUID
}

type ExerciseWithChoices struct {
	Choices  []entities.ExerciseChoiceEntity
	Exercise entities.ExerciseEntity
	Category string
	PrevId   string
	NextId   string
}

type ExerciseChoice struct {
	Id         uuid.UUID
	Value      string
	IsSolution bool
}

type ExerciseCategory struct {
	Iid      int
	Category string
}

type ExerciseUpsertForm struct {
	Id           string
	ProblemText  string
	SolutionText string
	Choices      []ExerciseChoice
	Category     ExerciseCategory
	ExameYear    string
	ExameFase    string
	CreatedBy    uuid.UUID
	UpdatedBy    uuid.UUID
}

type ExerciseAnswer struct {
	Id       string
	ChoiceId string
}
