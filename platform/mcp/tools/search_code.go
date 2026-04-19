package tools

import (
	"context"
	"fmt"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	"github.com/mark3labs/mcp-go/mcp"
)

type SearchCodeUseCase interface {
	Execute(ctx context.Context, request distillapp.SearchCodeRequest) (distillapp.SearchCodeResult, error)
}

type SearchCode struct {
	useCase SearchCodeUseCase
}

func NewSearchCode(useCase SearchCodeUseCase) SearchCode {
	return SearchCode{useCase: useCase}
}

func (t SearchCode) Definition() mcp.Tool {
	return mcp.NewTool("search_code",
		mcp.WithDescription("Search repository code and return compact matches for next reasoning step"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("mode",
			mcp.Required(),
			mcp.Description("Search mode: text | regex | symbol | path"),
			mcp.Enum(distillapp.SearchModeText, distillapp.SearchModeRegex, distillapp.SearchModeSymbol, distillapp.SearchModePath),
		),
		mcp.WithString("question", mcp.Required(), mcp.Description("Output contract for the distilled result")),
		mcp.WithArray("scope", mcp.WithStringItems(), mcp.Description("Optional scope globs")),
		mcp.WithNumber("max_results", mcp.Description("Hard limit for returned candidates")),
		mcp.WithNumber("context_lines", mcp.Description("Context lines per match")),
	)
}

func (t SearchCode) Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := t.useCase.Execute(ctx, distillapp.SearchCodeRequest{
		Query:        mcp.ParseString(request, "query", ""),
		Mode:         mcp.ParseString(request, "mode", ""),
		Question:     mcp.ParseString(request, "question", ""),
		Scope:        parseScopeArgument(request),
		MaxResults:   mcp.ParseInt(request, "max_results", distillapp.DefaultSearchCodeMaxResults),
		ContextLines: mcp.ParseInt(request, "context_lines", distillapp.DefaultSearchCodeContextLines),
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result.Output), nil
}

func parseScopeArgument(request mcp.CallToolRequest) []string {
	argument := mcp.ParseArgument(request, "scope", []any{})
	if argument == nil {
		return nil
	}

	scope, ok := argument.([]any)
	if !ok {
		scopeStrings, castOK := argument.([]string)
		if castOK {
			return scopeStrings
		}
		return nil
	}

	result := make([]string, 0, len(scope))
	for _, item := range scope {
		value, valueOK := item.(string)
		if !valueOK {
			result = append(result, fmt.Sprint(item))
			continue
		}
		result = append(result, value)
	}

	return result
}
