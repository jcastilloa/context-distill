package distillation

import (
	"fmt"
	"sort"
	"strings"
)

const maxSearchSnippetLength = 240

func compactSearchMatches(request SearchCodeRequest, matches []SearchMatch) []SearchMatch {
	unique := dedupeSearchMatches(matches)
	trimmed := trimSearchSnippets(unique, maxSearchSnippetLength)
	prioritized := prioritizeSearchMatches(request.Mode, trimmed)
	return applySearchResultLimit(request.MaxResults, prioritized)
}

func dedupeSearchMatches(matches []SearchMatch) []SearchMatch {
	seen := make(map[string]struct{}, len(matches))
	result := make([]SearchMatch, 0, len(matches))

	for _, match := range matches {
		key := searchMatchKey(match)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		result = append(result, match)
	}

	return result
}

func searchMatchKey(match SearchMatch) string {
	if match.Line <= 0 {
		return match.File
	}

	return fmt.Sprintf("%s:%d", match.File, match.Line)
}

func trimSearchSnippets(matches []SearchMatch, maxLen int) []SearchMatch {
	trimmed := make([]SearchMatch, 0, len(matches))
	for _, match := range matches {
		copyMatch := match
		copyMatch.Snippet = trimSnippet(copyMatch.Snippet, maxLen)
		trimmed = append(trimmed, copyMatch)
	}

	return trimmed
}

func trimSnippet(snippet string, maxLen int) string {
	normalized := strings.TrimSpace(snippet)
	if maxLen <= 0 || len(normalized) <= maxLen {
		return normalized
	}

	if maxLen <= 3 {
		return normalized[:maxLen]
	}

	return normalized[:maxLen-3] + "..."
}

func prioritizeSearchMatches(mode string, matches []SearchMatch) []SearchMatch {
	if mode != SearchModeSymbol {
		return matches
	}

	sorted := append([]SearchMatch(nil), matches...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return symbolKindPriority(sorted[i].Kind) < symbolKindPriority(sorted[j].Kind)
	})

	return sorted
}

func symbolKindPriority(kind string) int {
	switch kind {
	case SearchMatchKindDefinition:
		return 0
	case SearchMatchKindUsage:
		return 1
	case SearchMatchKindMatch:
		return 2
	default:
		return 3
	}
}

func applySearchResultLimit(limit int, matches []SearchMatch) []SearchMatch {
	if limit <= 0 || len(matches) <= limit {
		return matches
	}
	return matches[:limit]
}
