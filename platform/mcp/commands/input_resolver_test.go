package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInputReturnsFlagValueWhenProvided(t *testing.T) {
	result, err := resolveInput("flag data", strings.NewReader("stdin data"))

	require.NoError(t, err)
	assert.Equal(t, "flag data", result)
}

func TestResolveInputReadsStdinWhenFlagIsEmpty(t *testing.T) {
	result, err := resolveInput("", strings.NewReader("piped content"))

	require.NoError(t, err)
	assert.Equal(t, "piped content", result)
}

func TestResolveInputReadsStdinWhenFlagIsDash(t *testing.T) {
	result, err := resolveInput("-", strings.NewReader("piped via dash"))

	require.NoError(t, err)
	assert.Equal(t, "piped via dash", result)
}

func TestResolveInputTrimsTrailingNewlineFromStdin(t *testing.T) {
	result, err := resolveInput("", strings.NewReader("content\n"))

	require.NoError(t, err)
	assert.Equal(t, "content", result)
}

func TestResolveInputPreservesInternalNewlinesFromStdin(t *testing.T) {
	result, err := resolveInput("", strings.NewReader("line1\nline2\nline3\n"))

	require.NoError(t, err)
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestResolveInputReturnsErrorWhenStdinIsEmpty(t *testing.T) {
	_, err := resolveInput("", strings.NewReader(""))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}
