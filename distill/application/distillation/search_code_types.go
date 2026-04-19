package distillation

import (
	"fmt"
	"strings"
)

const (
	SearchModeText   = "text"
	SearchModeRegex  = "regex"
	SearchModeSymbol = "symbol"
	SearchModePath   = "path"

	SearchMatchKindMatch      = "match"
	SearchMatchKindDefinition = "definition"
	SearchMatchKindUsage      = "usage"
	SearchMatchKindPath       = "path"

	DefaultSearchCodeMaxResults   = 20
	DefaultSearchCodeContextLines = 2
)

type SearchCodeRequest struct {
	Query        string
	Mode         string
	Question     string
	Scope        []string
	MaxResults   int
	ContextLines int
}

type SearchMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line,omitempty"`
	Snippet string `json:"snippet,omitempty"`
	Kind    string `json:"kind,omitempty"`
}

type SearchCodeResult struct {
	RawMatches   []SearchMatch `json:"raw_matches,omitempty"`
	Output       string        `json:"output"`
	UsedFallback bool          `json:"used_fallback,omitempty"`
}

func (r SearchCodeRequest) WithDefaults() SearchCodeRequest {
	withDefaults := SearchCodeRequest{
		Query:        strings.TrimSpace(r.Query),
		Mode:         strings.ToLower(strings.TrimSpace(r.Mode)),
		Question:     strings.TrimSpace(r.Question),
		Scope:        normalizeScope(r.Scope),
		MaxResults:   r.MaxResults,
		ContextLines: r.ContextLines,
	}

	if withDefaults.MaxResults == 0 {
		withDefaults.MaxResults = DefaultSearchCodeMaxResults
	}

	return withDefaults
}

func (r SearchCodeRequest) Validate() error {
	if strings.TrimSpace(r.Query) == "" {
		return ErrQueryRequired
	}
	if strings.TrimSpace(r.Question) == "" {
		return ErrQuestionRequired
	}

	mode := strings.ToLower(strings.TrimSpace(r.Mode))
	if mode == "" {
		return ErrModeRequired
	}
	if !isSupportedSearchMode(mode) {
		return fmt.Errorf("%w: %s", ErrUnsupportedSearchMode, mode)
	}

	if r.MaxResults < 0 {
		return ErrInvalidMaxResults
	}
	if r.ContextLines < 0 {
		return ErrInvalidContextLines
	}

	return nil
}

func isSupportedSearchMode(mode string) bool {
	switch mode {
	case SearchModeText, SearchModeRegex, SearchModeSymbol, SearchModePath:
		return true
	default:
		return false
	}
}

func normalizeScope(scope []string) []string {
	normalized := make([]string, 0, len(scope))
	for _, raw := range scope {
		candidate := strings.TrimSpace(raw)
		if candidate == "" {
			continue
		}
		normalized = append(normalized, candidate)
	}

	return normalized
}
