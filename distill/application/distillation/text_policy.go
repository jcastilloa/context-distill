package distillation

import (
	"regexp"
	"strings"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
)

const (
	promptTailWindow       = 256
	longSourceThreshold    = 1024
	longSourceMaxRatio     = 0.8
	shortSourceMaxOverhead = 40
	signatureMaxLines      = 24
)

var (
	ansiPattern           = regexp.MustCompile(`\x1b(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
	trailingWhitespaceEOL = regexp.MustCompile(`[ \t]+\n`)
	multiNewlinePattern   = regexp.MustCompile(`\n{3,}`)
	promptPattern         = regexp.MustCompile(`(?i)(?:\[[Yy]/[Nn]\]|\[[Nn]/[Yy]\]|\([Yy]/[Nn]\)|\([Nn]/[Yy]\)|password:|passphrase:|continue\?|proceed\?)\s*$`)
	numberPattern         = regexp.MustCompile(`\b\d+\b`)
	hexPattern            = regexp.MustCompile(`[0-9a-f]{7,}`)
	spacePattern          = regexp.MustCompile(`\s+`)
)

type TextPolicy struct{}

func NewTextPolicy() distilldomain.TextPolicy {
	return TextPolicy{}
}

func (TextPolicy) NormalizeForModel(input string) string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = ansiPattern.ReplaceAllString(normalized, "")
	normalized = trailingWhitespaceEOL.ReplaceAllString(normalized, "\n")
	normalized = multiNewlinePattern.ReplaceAllString(normalized, "\n\n")
	return strings.TrimSpace(normalized)
}

func (TextPolicy) HasPromptLikeTail(input string) bool {
	tail := input
	if len(tail) > promptTailWindow {
		tail = tail[len(tail)-promptTailWindow:]
	}

	return promptPattern.MatchString(strings.TrimRight(tail, " \t\n\r"))
}

func (TextPolicy) HasRedrawSignal(input string) bool {
	return strings.Contains(input, "\r") ||
		strings.Contains(input, "\x1b[2J") ||
		strings.Contains(input, "\x1bc")
}

func (p TextPolicy) StructuralSimilarity(left, right string) float64 {
	leftSignature := p.structuralSignature(left)
	rightSignature := p.structuralSignature(right)
	if len(leftSignature) == 0 || len(rightSignature) == 0 {
		return 0
	}

	leftSet := make(map[string]struct{}, len(leftSignature))
	rightSet := make(map[string]struct{}, len(rightSignature))

	for _, value := range leftSignature {
		leftSet[value] = struct{}{}
	}
	for _, value := range rightSignature {
		rightSet[value] = struct{}{}
	}

	overlap := 0
	for value := range leftSet {
		if _, exists := rightSet[value]; exists {
			overlap++
		}
	}

	return (2 * float64(overlap)) / (float64(len(leftSet)) + float64(len(rightSet)))
}

func (p TextPolicy) LooksLikeBadDistillation(source, candidate string) bool {
	normalizedSource := p.NormalizeForModel(source)
	normalizedCandidate := p.NormalizeForModel(candidate)
	if normalizedCandidate == "" {
		return true
	}

	lowerCandidate := strings.ToLower(normalizedCandidate)
	if strings.Contains(lowerCandidate, "please provide") ||
		strings.Contains(lowerCandidate, "wish summarized") ||
		strings.Contains(lowerCandidate, "provided command output") {
		return true
	}

	if len(normalizedSource) >= longSourceThreshold {
		return float64(len(normalizedCandidate)) >= float64(len(normalizedSource))*longSourceMaxRatio
	}

	if len(normalizedSource) > 0 {
		return normalizedCandidate == normalizedSource ||
			len(normalizedCandidate) > len(normalizedSource)+shortSourceMaxOverhead
	}

	return false
}

func (TextPolicy) EnsureTrailingNewline(input string) string {
	if strings.HasSuffix(input, "\n") {
		return input
	}
	return input + "\n"
}

func (p TextPolicy) structuralSignature(input string) []string {
	lines := strings.Split(p.NormalizeForModel(input), "\n")
	signature := make([]string, 0, len(lines))

	for _, line := range lines {
		normalizedLine := strings.ToLower(line)
		normalizedLine = numberPattern.ReplaceAllString(normalizedLine, "#")
		normalizedLine = hexPattern.ReplaceAllString(normalizedLine, "<hex>")
		normalizedLine = spacePattern.ReplaceAllString(normalizedLine, " ")
		normalizedLine = strings.TrimSpace(normalizedLine)

		if normalizedLine != "" {
			signature = append(signature, normalizedLine)
		}
	}

	if len(signature) <= signatureMaxLines {
		return signature
	}

	return append([]string(nil), signature[:signatureMaxLines]...)
}
