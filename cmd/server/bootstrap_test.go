package main

import "testing"

func TestShouldUseSetupConfigLoaderReturnsTrueForConfigUIFlag(t *testing.T) {
	testCases := [][]string{
		{"--config-ui"},
		{"--config-ui=true"},
		{"--transport", "stdio", "--config-ui"},
	}

	for _, args := range testCases {
		if !shouldUseSetupConfigLoader(args) {
			t.Fatalf("expected setup loader for args: %v", args)
		}
	}
}

func TestShouldUseSetupConfigLoaderReturnsFalseWithoutFlag(t *testing.T) {
	testCases := [][]string{
		nil,
		{},
		{"--transport", "stdio"},
		{"version"},
	}

	for _, args := range testCases {
		if shouldUseSetupConfigLoader(args) {
			t.Fatalf("did not expect setup loader for args: %v", args)
		}
	}
}
