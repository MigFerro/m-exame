package data

import (
	"github.com/MigFerro/exame/entities"
	"github.com/google/uuid"
)

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
}

type ExerciseChoice struct {
	Value      string
	IsSolution bool
}

type ExerciseCategory struct {
	Iid      int
	Category string
}

type ExerciseCreationForm struct {
	ProblemText string
	Choices     []ExerciseChoice
	Category    ExerciseCategory
	ExameYear   string
	ExameFase   string
	CreatedBy   uuid.UUID
}
