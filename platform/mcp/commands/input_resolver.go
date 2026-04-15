package commands

import (
	"fmt"
	"io"
	"strings"
)

const stdinMarker = "-"

func resolveInput(flagValue string, stdin io.Reader) (string, error) {
	if flagValue != "" && flagValue != stdinMarker {
		return flagValue, nil
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("reading input from stdin: %w", err)
	}

	content := strings.TrimRight(string(data), "\n")
	if content == "" {
		return "", fmt.Errorf("no input provided: use --input flag or pipe data via stdin")
	}

	return content, nil
}
