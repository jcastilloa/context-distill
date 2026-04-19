package search

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
)

const rgMaxFileSize = "1M"

var excludedGlobs = []string{"!.git/**", "!node_modules/**", "!vendor/**"}

type CommandRunner interface {
	Run(ctx context.Context, workDir, name string, args []string) ([]byte, error)
}

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type RipgrepSearcher struct {
	WorkDir string
	Runner  CommandRunner
	Reader  FileReader
}

func NewRipgrepSearcher(workDir string) *RipgrepSearcher {
	if strings.TrimSpace(workDir) == "" {
		workDir = "."
	}

	return &RipgrepSearcher{
		WorkDir: workDir,
		Runner:  osCommandRunner{},
		Reader:  osFileReader{},
	}
}

func (s *RipgrepSearcher) Search(ctx context.Context, request distillapp.SearchCodeRequest) ([]distillapp.SearchMatch, error) {
	switch request.Mode {
	case distillapp.SearchModeText, distillapp.SearchModeRegex, distillapp.SearchModeSymbol:
		return s.searchByContent(ctx, request)
	case distillapp.SearchModePath:
		return s.searchByPath(ctx, request)
	default:
		return nil, fmt.Errorf("%w: %s", distillapp.ErrUnsupportedSearchMode, request.Mode)
	}
}

func (s *RipgrepSearcher) searchByContent(ctx context.Context, request distillapp.SearchCodeRequest) ([]distillapp.SearchMatch, error) {
	args := buildContentSearchArgs(request)
	output, err := s.Runner.Run(ctx, s.WorkDir, "rg", args)
	if err != nil {
		if isNoMatchError(err) {
			return []distillapp.SearchMatch{}, nil
		}
		return nil, fmt.Errorf("run rg content search: %w", err)
	}

	matches, err := parseRipgrepJSONMatches(output)
	if err != nil {
		return nil, err
	}

	if request.Mode == distillapp.SearchModeSymbol {
		classifySymbolMatches(matches, request.Query)
	}

	if request.ContextLines > 0 {
		enrichMatchesWithContext(matches, request.ContextLines, s.WorkDir, s.Reader)
	}

	return matches, nil
}

func (s *RipgrepSearcher) searchByPath(ctx context.Context, request distillapp.SearchCodeRequest) ([]distillapp.SearchMatch, error) {
	args := buildPathSearchArgs(request)
	output, err := s.Runner.Run(ctx, s.WorkDir, "rg", args)
	if err != nil {
		if isNoMatchError(err) {
			return []distillapp.SearchMatch{}, nil
		}
		return nil, fmt.Errorf("run rg path search: %w", err)
	}

	paths := strings.Split(strings.TrimSpace(string(output)), "\n")
	query := strings.ToLower(request.Query)
	matches := make([]distillapp.SearchMatch, 0, len(paths))
	for _, path := range paths {
		candidate := strings.TrimSpace(path)
		if candidate == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(candidate), query) {
			continue
		}

		matches = append(matches, distillapp.SearchMatch{File: candidate, Kind: distillapp.SearchMatchKindPath})
	}

	return matches, nil
}

func buildContentSearchArgs(request distillapp.SearchCodeRequest) []string {
	args := []string{"--json", "--line-number", "--with-filename", "--color", "never", "--max-filesize", rgMaxFileSize}
	if request.Mode == distillapp.SearchModeText || request.Mode == distillapp.SearchModeSymbol {
		args = append(args, "--fixed-strings")
	}

	args = appendGlobFilters(args, request.Scope)
	args = append(args, request.Query, ".")
	return args
}

func buildPathSearchArgs(request distillapp.SearchCodeRequest) []string {
	args := []string{"--files", "--color", "never"}
	args = appendGlobFilters(args, request.Scope)
	return args
}

func appendGlobFilters(args []string, scope []string) []string {
	for _, glob := range excludedGlobs {
		args = append(args, "--glob", glob)
	}
	for _, candidate := range scope {
		args = append(args, "--glob", candidate)
	}

	return args
}

func parseRipgrepJSONMatches(output []byte) ([]distillapp.SearchMatch, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	matches := make([]distillapp.SearchMatch, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		event := ripgrepEvent{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("parse rg json line %q: %w", line, err)
		}

		if event.Type != "match" {
			continue
		}

		match := distillapp.SearchMatch{
			File:    strings.TrimSpace(event.Data.Path.Text),
			Line:    event.Data.LineNumber,
			Snippet: strings.TrimSpace(strings.ReplaceAll(event.Data.Lines.Text, "\n", "")),
			Kind:    distillapp.SearchMatchKindMatch,
		}
		if match.File == "" {
			continue
		}

		matches = append(matches, match)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan rg output: %w", err)
	}

	return matches, nil
}

func classifySymbolMatches(matches []distillapp.SearchMatch, symbol string) {
	for index := range matches {
		matches[index].Kind = classifySymbolLine(matches[index].Snippet, symbol)
	}
}

func classifySymbolLine(line, symbol string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return distillapp.SearchMatchKindUsage
	}

	if strings.Contains(trimmed, "func "+symbol+"(") ||
		strings.Contains(trimmed, "type "+symbol+" ") ||
		strings.Contains(trimmed, "type "+symbol+" struct") ||
		strings.Contains(trimmed, "var "+symbol+" ") ||
		strings.Contains(trimmed, "const "+symbol+" ") {
		return distillapp.SearchMatchKindDefinition
	}

	if strings.HasPrefix(trimmed, "func ") && strings.Contains(trimmed, ") "+symbol+"(") {
		return distillapp.SearchMatchKindDefinition
	}

	return distillapp.SearchMatchKindUsage
}

func enrichMatchesWithContext(matches []distillapp.SearchMatch, contextLines int, workDir string, reader FileReader) {
	if contextLines <= 0 {
		return
	}

	cache := map[string][]string{}
	for index := range matches {
		if matches[index].Line <= 0 {
			continue
		}

		lines, ok := cache[matches[index].File]
		if !ok {
			content, err := reader.ReadFile(filepath.Join(workDir, matches[index].File))
			if err != nil {
				continue
			}
			lines = splitLines(string(content))
			cache[matches[index].File] = lines
		}

		snippet := buildContextSnippet(lines, matches[index].Line, contextLines)
		if snippet != "" {
			matches[index].Snippet = snippet
		}
	}
}

func splitLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	return strings.Split(normalized, "\n")
}

func buildContextSnippet(lines []string, lineNumber, contextLines int) string {
	if lineNumber <= 0 || lineNumber > len(lines) {
		return ""
	}

	start := max(lineNumber-contextLines, 1)
	end := min(lineNumber+contextLines, len(lines))

	selected := make([]string, 0, end-start+1)
	for idx := start; idx <= end; idx++ {
		selected = append(selected, strconv.Itoa(idx)+": "+strings.TrimSpace(lines[idx-1]))
	}

	return strings.TrimSpace(strings.Join(selected, "\n"))
}

func isNoMatchError(err error) bool {
	var exitCoder interface{ ExitCode() int }
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode() == 1
	}
	return false
}

type osCommandRunner struct{}

func (osCommandRunner) Run(ctx context.Context, workDir, name string, args []string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = workDir
	return command.CombinedOutput()
}

type osFileReader struct{}

func (osFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type ripgrepEvent struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
	} `json:"data"`
}
