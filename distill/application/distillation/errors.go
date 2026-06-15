package distillation

import "errors"

var (
	ErrQuestionRequired      = errors.New("question is required")
	ErrInputRequired         = errors.New("input is required")
	ErrMCPOutputRequired     = errors.New("mcp output is required")
	ErrMCPInvocationRequired = errors.New("either mcp output or server invocation is required")
	ErrMCPServerCommandReq   = errors.New("server command is required when mcp output is omitted")
	ErrMCPToolNameRequired   = errors.New("tool name is required when invoking an mcp server")
	ErrPreviousCycleRequired = errors.New("previous cycle is required")
	ErrCurrentCycleRequired  = errors.New("current cycle is required")
	ErrQueryRequired         = errors.New("query is required")
	ErrModeRequired          = errors.New("mode is required")
	ErrUnsupportedSearchMode = errors.New("unsupported mode")
	ErrInvalidMaxResults     = errors.New("max-results must be >= 0")
	ErrInvalidContextLines   = errors.New("context-lines must be >= 0")
)
