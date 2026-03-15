package distillation

import "errors"

var (
	ErrQuestionRequired      = errors.New("question is required")
	ErrInputRequired         = errors.New("input is required")
	ErrPreviousCycleRequired = errors.New("previous cycle is required")
	ErrCurrentCycleRequired  = errors.New("current cycle is required")
)
