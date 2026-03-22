package di

import (
	"fmt"

	distillApp "github.com/jcastilloa/context-distill/distill/application/distillation"
	distillDomain "github.com/jcastilloa/context-distill/distill/domain"
	"github.com/jcastilloa/context-distill/platform/configui"
	"github.com/jcastilloa/context-distill/platform/mcp/commands"
	mcpserver "github.com/jcastilloa/context-distill/platform/mcp/server"
	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	aiDomain "github.com/jcastilloa/context-distill/shared/ai/domain"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"

	"github.com/sarulabs/di"
)

type Container struct {
	aiRepository          aiDomain.AIRepository
	distillProviderConfig aiDomain.ProviderConfig
	serviceName           string
	serviceCfg            configDomain.ServiceConfig
}

func New(
	aiRepository aiDomain.AIRepository,
	distillProviderConfig aiDomain.ProviderConfig,
	serviceName string,
	serviceCfg configDomain.ServiceConfig,
) *Container {
	return &Container{
		aiRepository:          aiRepository,
		distillProviderConfig: distillProviderConfig,
		serviceName:           serviceName,
		serviceCfg:            serviceCfg,
	}
}

func (c *Container) Build() (*di.Container, error) {
	builder, err := di.NewBuilder()
	if err != nil {
		return nil, fmt.Errorf("create builder: %w", err)
	}

	err = builder.Add(
		di.Def{
			Name:  OpenAIRepositoryLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return c.aiRepository, nil
			},
		},
		di.Def{
			Name:  ConfigUIRepositoryLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return configui.NewFileRepository(), nil
			},
		},
		di.Def{
			Name:  ConfigUIRunnerLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				repository := ctn.Get(ConfigUIRepositoryLabel).(configui.Repository)
				return configui.NewTViewRunner(repository), nil
			},
		},
		di.Def{
			Name:  DistillPromptBuilderLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return distillApp.NewPromptBuilder(), nil
			},
		},
		di.Def{
			Name:  DistillTextPolicyLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return distillApp.NewTextPolicy(), nil
			},
		},
		di.Def{
			Name:  DistillSummarizerRepositoryLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				aiRepository := ctn.Get(OpenAIRepositoryLabel).(aiDomain.AIRepository)
				return newDistillSummarizerRepository(c.distillProviderConfig, aiRepository)
			},
		},
		di.Def{
			Name:  DistillBatchUseCaseLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				promptBuilder := ctn.Get(DistillPromptBuilderLabel).(distillDomain.PromptBuilder)
				textPolicy := ctn.Get(DistillTextPolicyLabel).(distillDomain.TextPolicy)
				summarizer := ctn.Get(DistillSummarizerRepositoryLabel).(distillDomain.SummarizerRepository)
				return distillApp.NewDistillBatchUseCase(promptBuilder, textPolicy, summarizer), nil
			},
		},
		di.Def{
			Name:  DistillWatchUseCaseLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				promptBuilder := ctn.Get(DistillPromptBuilderLabel).(distillDomain.PromptBuilder)
				textPolicy := ctn.Get(DistillTextPolicyLabel).(distillDomain.TextPolicy)
				summarizer := ctn.Get(DistillSummarizerRepositoryLabel).(distillDomain.SummarizerRepository)
				return distillApp.NewDistillWatchUseCase(promptBuilder, textPolicy, summarizer), nil
			},
		},
		di.Def{
			Name:  DistillBatchToolLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				useCase := ctn.Get(DistillBatchUseCaseLabel).(*distillApp.DistillBatchUseCase)
				return tools.NewDistillBatch(useCase), nil
			},
		},
		di.Def{
			Name:  DistillWatchToolLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				useCase := ctn.Get(DistillWatchUseCaseLabel).(*distillApp.DistillWatchUseCase)
				return tools.NewDistillWatch(useCase), nil
			},
		},
		di.Def{
			Name:  MCPServerLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				return mcpserver.New(c.serviceName, c.serviceCfg.Version), nil
			},
		},
		di.Def{
			Name:  commands.RootCommandLabel,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				server := ctn.Get(MCPServerLabel).(*mcpserver.Server)
				configUIRunner := ctn.Get(ConfigUIRunnerLabel).(commands.ConfigUIRunner)
				distillBatchTool := ctn.Get(DistillBatchToolLabel).(tools.DistillBatch)
				distillWatchTool := ctn.Get(DistillWatchToolLabel).(tools.DistillWatch)
				distillBatchUseCase := ctn.Get(DistillBatchUseCaseLabel).(*distillApp.DistillBatchUseCase)
				distillWatchUseCase := ctn.Get(DistillWatchUseCaseLabel).(*distillApp.DistillWatchUseCase)
				return commands.NewRunner(
					c.serviceName,
					c.serviceCfg,
					server,
					configUIRunner,
					distillBatchTool,
					distillWatchTool,
					distillBatchUseCase,
					distillWatchUseCase,
				), nil
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("register dependencies: %w", err)
	}

	container := builder.Build()
	return &container, nil
}
