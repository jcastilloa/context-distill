package distillation

import "context"

type fakePromptBuilder struct {
	batchPrompt string
	watchPrompt string

	batchQuestion string
	batchInput    string

	watchQuestion string
	watchPrevious string
	watchCurrent  string
}

func (f *fakePromptBuilder) BuildBatchPrompt(question, input string) string {
	f.batchQuestion = question
	f.batchInput = input
	if f.batchPrompt != "" {
		return f.batchPrompt
	}
	return "batch-prompt"
}

func (f *fakePromptBuilder) BuildWatchPrompt(question, previousCycle, currentCycle string) string {
	f.watchQuestion = question
	f.watchPrevious = previousCycle
	f.watchCurrent = currentCycle
	if f.watchPrompt != "" {
		return f.watchPrompt
	}
	return "watch-prompt"
}

type fakeTextPolicy struct {
	normalizedValue   string
	normalizedByInput map[string]string
	badDistillation   bool

	normalizeCalls []string
	badSource      string
	badCandidate   string
	ensuredValue   string
}

func (f *fakeTextPolicy) NormalizeForModel(input string) string {
	f.normalizeCalls = append(f.normalizeCalls, input)
	if f.normalizedByInput != nil {
		if value, exists := f.normalizedByInput[input]; exists {
			return value
		}
	}
	if f.normalizedValue != "" {
		return f.normalizedValue
	}
	return input
}

func (f *fakeTextPolicy) HasPromptLikeTail(string) bool {
	return false
}

func (f *fakeTextPolicy) HasRedrawSignal(string) bool {
	return false
}

func (f *fakeTextPolicy) StructuralSimilarity(string, string) float64 {
	return 0
}

func (f *fakeTextPolicy) LooksLikeBadDistillation(source, candidate string) bool {
	f.badSource = source
	f.badCandidate = candidate
	return f.badDistillation
}

func (f *fakeTextPolicy) EnsureTrailingNewline(input string) string {
	f.ensuredValue = input
	return input + "\n"
}

type fakeSummarizerRepository struct {
	batchSummary string
	watchSummary string

	batchErr error
	watchErr error

	batchPrompt string
	watchPrompt string
	batchCalls  int
	watchCalls  int
}

func (f *fakeSummarizerRepository) SummarizeBatch(ctx context.Context, prompt string) (string, error) {
	_ = ctx
	f.batchPrompt = prompt
	f.batchCalls++
	if f.batchErr != nil {
		return "", f.batchErr
	}
	return f.batchSummary, nil
}

func (f *fakeSummarizerRepository) SummarizeWatch(ctx context.Context, prompt string) (string, error) {
	_ = ctx
	f.watchPrompt = prompt
	f.watchCalls++
	if f.watchErr != nil {
		return "", f.watchErr
	}
	return f.watchSummary, nil
}
