package shared

type ExerciseChoice struct {
	Value      string
	IsSolution bool
}

type ExerciseCategory struct {
	Iid      int
	Category string
}

type ExerciseFormResponse struct {
	ProblemText string
	Choices     []ExerciseChoice
	Category    ExerciseCategory
	ExameYear   string
	ExameFase   string
}
