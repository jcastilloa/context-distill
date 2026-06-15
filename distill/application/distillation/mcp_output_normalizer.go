package distillation

import (
	"encoding/json"
	"strings"
)

func normalizeMCPOutput(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return trimmed
	}

	text := collectMCPText(payload)
	if text != "" {
		return text
	}

	return prettyJSON(payload, trimmed)
}

func collectMCPText(payload any) string {
	parts := collectMCPTextParts(payload)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}

func collectMCPTextParts(payload any) []string {
	switch value := payload.(type) {
	case map[string]any:
		return collectMCPMapText(value)
	case []any:
		return collectMCPArrayText(value)
	case string:
		return []string{strings.TrimSpace(value)}
	default:
		return nil
	}
}

func collectMCPMapText(value map[string]any) []string {
	text, ok := value["text"].(string)
	if ok && strings.TrimSpace(text) != "" {
		return []string{strings.TrimSpace(text)}
	}

	parts := collectMCPTextParts(value["content"])
	if len(parts) > 0 {
		return parts
	}

	structuredParts := collectMCPStructuredContent(value["structuredContent"])
	if len(structuredParts) > 0 {
		return structuredParts
	}

	return nil
}

func collectMCPArrayText(value []any) []string {
	parts := make([]string, 0, len(value))
	for _, item := range value {
		parts = append(parts, collectMCPTextParts(item)...)
	}
	return parts
}

func collectMCPStructuredContent(value any) []string {
	if value == nil {
		return nil
	}

	pretty := prettyJSON(value, "")
	if pretty == "" {
		return nil
	}

	return []string{pretty}
}

func prettyJSON(payload any, fallback string) string {
	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fallback
	}

	return string(formatted)
}
