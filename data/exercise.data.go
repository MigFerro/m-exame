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
	ExerciseId string
	IsSolution bool
	NextId     string
	At         string
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
	Id          string
	ProblemText string
	Choices     []ExerciseChoice
	Category    ExerciseCategory
	ExameYear   string
	ExameFase   string
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}
