package domain

// TextPolicy encapsulates text heuristics used by distillation workflows.
type TextPolicy interface {
	NormalizeForModel(input string) string
	HasPromptLikeTail(input string) bool
	HasRedrawSignal(input string) bool
	StructuralSimilarity(left, right string) float64
	LooksLikeBadDistillation(source, candidate string) bool
	EnsureTrailingNewline(input string) string
}
