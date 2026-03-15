package di

import "testing"

func TestIsOpenAICompatibleProviderSupportsConfiguredProviders(t *testing.T) {
	testCases := []string{
		"openai",
		"openai-compatible",
		"openrouter",
		"lmstudio",
		"jan",
		"localai",
		"vllm",
		"sglang",
		"llama.cpp",
		"mlx-lm",
		"docker-model-runner",
	}

	for _, provider := range testCases {
		if !IsOpenAICompatibleProvider(provider) {
			t.Fatalf("expected provider %q to be openai-compatible", provider)
		}
	}
}

func TestIsOpenAICompatibleProviderRejectsUnknownProvider(t *testing.T) {
	if IsOpenAICompatibleProvider("custom-provider") {
		t.Fatalf("expected custom-provider to be rejected")
	}
}

func TestNormalizeProviderNameSupportsAliases(t *testing.T) {
	testCases := map[string]string{
		"OpenAI Compatible": "openai-compatible",
		"openai_compatible": "openai-compatible",
		"llamacpp":          "llama.cpp",
		"DMR":               "docker-model-runner",
	}

	for input, expected := range testCases {
		actual := NormalizeProviderName(input)
		if actual != expected {
			t.Fatalf("expected %q for %q, got %q", expected, input, actual)
		}
	}
}
