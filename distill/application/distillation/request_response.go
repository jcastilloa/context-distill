package distillation

type DistillBatchRequest struct {
	Question string
	Input    string
}

type DistillBatchResult struct {
	Output       string
	UsedFallback bool
}

type DistillMCPOutputRequest struct {
	Question      string
	ToolName      string
	Output        string
	ServerCommand string
	ServerArgs    []string
	ToolArguments any
}

type DistillMCPOutputResult struct {
	Output       string
	UsedFallback bool
}

type DistillMCPToolCallRequest struct {
	ServerCommand string
	ServerArgs    []string
	ToolName      string
	ToolArguments any
}

type DistillMCPToolCallResult struct {
	Output string
}

type DistillWatchRequest struct {
	Question      string
	PreviousCycle string
	CurrentCycle  string
}

type DistillWatchResult struct {
	Output       string
	UsedFallback bool
}
