package domain

// PromptBuilder builds provider-ready prompts from distillation inputs.
type PromptBuilder interface {
	BuildBatchPrompt(question, input string) string
	BuildWatchPrompt(question, previousCycle, currentCycle string) string
}
