package distillation

type DistillBatchRequest struct {
	Question string
	Input    string
}

type DistillBatchResult struct {
	Output       string
	UsedFallback bool
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
