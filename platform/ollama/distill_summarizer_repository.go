package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
)

const (
	defaultBaseURL = "http://127.0.0.1:11434"
	defaultModel   = "qwen3.5:2b"
	defaultTimeout = 90 * time.Second
)

type DistillSummarizerRepository struct {
	client *http.Client
	cfg    aiDomain.ProviderConfig
}

type requestPayload struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Think   bool           `json:"think"`
	Options requestOptions `json:"options"`
}

type requestOptions struct {
	Temperature float64 `json:"temperature"`
	NumPredict  int     `json:"num_predict"`
}

type responsePayload struct {
	Response string `json:"response"`
}

func NewDistillSummarizerRepository(cfg aiDomain.ProviderConfig, client *http.Client) distilldomain.SummarizerRepository {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = defaultModel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}

	return &DistillSummarizerRepository{client: client, cfg: cfg}
}

func (r *DistillSummarizerRepository) SummarizeBatch(ctx context.Context, prompt string) (string, error) {
	return r.summarize(ctx, prompt)
}

func (r *DistillSummarizerRepository) SummarizeWatch(ctx context.Context, prompt string) (string, error) {
	return r.summarize(ctx, prompt)
}

func (r *DistillSummarizerRepository) summarize(ctx context.Context, prompt string) (string, error) {
	payload := requestPayload{
		Model:  r.cfg.Model,
		Prompt: prompt,
		Stream: false,
		Think:  r.cfg.Thinking,
		Options: requestOptions{
			Temperature: 0.1,
			NumPredict:  80,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	url := strings.TrimRight(r.cfg.BaseURL, "/") + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama request failed with %d", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read ollama response: %w", err)
	}

	var response responsePayload
	if err = json.Unmarshal(rawBody, &response); err != nil {
		return "", fmt.Errorf("ollama returned invalid JSON: %w", err)
	}

	output := strings.TrimSpace(response.Response)
	if output == "" {
		return "", fmt.Errorf("ollama returned empty response")
	}

	return output, nil
}
