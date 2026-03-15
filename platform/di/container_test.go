package di

import (
	"context"
	"testing"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	distilldomain "github.com/jcastilloa/context-distill/distill/domain"
	"github.com/jcastilloa/context-distill/platform/mcp/commands"
	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"
)

type fakeAIRepository struct{}

func (fakeAIRepository) GetAIResponse(context.Context, *aiDomain.Request) (*aiDomain.Response, error) {
	return aiDomain.NewResponse("ok", nil, "", "", "", nil, false, "")
}

func TestContainerBuildRegistersDistillDependencies(t *testing.T) {
	builder := New(
		fakeAIRepository{},
		aiDomain.ProviderConfig{ProviderName: "openai-compatible"},
		"context-distill",
		configDomain.ServiceConfig{Version: "test"},
	)
	container, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	if _, ok := (*container).Get(DistillPromptBuilderLabel).(distilldomain.PromptBuilder); !ok {
		t.Fatalf("distill prompt builder is not registered as domain interface")
	}

	if _, ok := (*container).Get(DistillTextPolicyLabel).(distilldomain.TextPolicy); !ok {
		t.Fatalf("distill text policy is not registered as domain interface")
	}

	if _, ok := (*container).Get(DistillSummarizerRepositoryLabel).(distilldomain.SummarizerRepository); !ok {
		t.Fatalf("distill summarizer repository is not registered as domain interface")
	}

	if _, ok := (*container).Get(DistillBatchUseCaseLabel).(*distillapp.DistillBatchUseCase); !ok {
		t.Fatalf("distill batch use case is not registered")
	}

	if _, ok := (*container).Get(DistillWatchUseCaseLabel).(*distillapp.DistillWatchUseCase); !ok {
		t.Fatalf("distill watch use case is not registered")
	}

	if _, ok := (*container).Get(DistillBatchToolLabel).(tools.DistillBatch); !ok {
		t.Fatalf("distill batch tool is not registered")
	}

	if _, ok := (*container).Get(DistillWatchToolLabel).(tools.DistillWatch); !ok {
		t.Fatalf("distill watch tool is not registered")
	}

	if _, ok := (*container).Get(ConfigUIRunnerLabel).(commands.ConfigUIRunner); !ok {
		t.Fatalf("config ui runner is not registered")
	}
}
