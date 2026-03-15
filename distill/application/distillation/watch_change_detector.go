package distillation

import (
	"regexp"
	"sort"
	"strings"
)

var (
	durationPattern = regexp.MustCompile(`\b\d+(?:\.\d+)?(?:ns|us|µs|ms|s|m|h)\b`)
	sizePattern     = regexp.MustCompile(`\b\d+(?:\.\d+)?(?:b|kb|mb|gb|tb|kib|mib|gib|tib)\b`)
	percentPattern  = regexp.MustCompile(`\b\d+(?:\.\d+)?%\b`)
	timePattern     = regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`)
	isoDatePattern  = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[tT]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:z|[+-]\d{2}:?\d{2})?\b`)
)

type WatchChangeDetector struct{}

func NewWatchChangeDetector() WatchChangeDetector {
	return WatchChangeDetector{}
}

func (WatchChangeDetector) HasRelevantChange(previousCycle, currentCycle string) bool {
	previousSignature := signatureWithoutNoise(previousCycle)
	currentSignature := signatureWithoutNoise(currentCycle)

	if len(previousSignature) != len(currentSignature) {
		return true
	}

	for index := range previousSignature {
		if previousSignature[index] != currentSignature[index] {
			return true
		}
	}

	return false
}

func signatureWithoutNoise(input string) []string {
	canonical := make([]string, 0)
	for _, line := range strings.Split(input, "\n") {
		normalizedLine := normalizeNoise(line)
		if normalizedLine == "" {
			continue
		}
		canonical = append(canonical, normalizedLine)
	}

	sort.Strings(canonical)
	return canonical
}

func normalizeNoise(line string) string {
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return ""
	}

	line = isoDatePattern.ReplaceAllString(line, "<timestamp>")
	line = timePattern.ReplaceAllString(line, "<time>")
	line = durationPattern.ReplaceAllString(line, "<duration>")
	line = sizePattern.ReplaceAllString(line, "<size>")
	line = percentPattern.ReplaceAllString(line, "<percent>")
	line = strings.Join(strings.Fields(line), " ")
	return line
}
