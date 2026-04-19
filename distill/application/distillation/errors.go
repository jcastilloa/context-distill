package distillation

import "errors"

var (
	ErrQuestionRequired      = errors.New("question is required")
	ErrInputRequired         = errors.New("input is required")
	ErrPreviousCycleRequired = errors.New("previous cycle is required")
	ErrCurrentCycleRequired  = errors.New("current cycle is required")
	ErrQueryRequired         = errors.New("query is required")
	ErrModeRequired          = errors.New("mode is required")
	ErrUnsupportedSearchMode = errors.New("unsupported mode")
	ErrInvalidMaxResults     = errors.New("max-results must be >= 0")
	ErrInvalidContextLines   = errors.New("context-lines must be >= 0")
)
