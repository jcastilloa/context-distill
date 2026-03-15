package main

import (
	"log"
	"os"

	"github.com/jcastilloa/context-distill/platform/config"
	containerdi "github.com/jcastilloa/context-distill/platform/di"
	"github.com/jcastilloa/context-distill/platform/mcp/commands"
	"github.com/jcastilloa/context-distill/platform/openai"
)

func main() {
	loadConfig := config.New
	if shouldUseSetupConfigLoader(os.Args[1:]) {
		loadConfig = config.NewForSetup
	}

	cfgRepo, err := loadConfig("context-distill")
	if err != nil {
		log.Fatal(err)
	}

	serviceCfg := cfgRepo.ServiceConfig()
	distillProviderCfg := cfgRepo.DistillProviderConfig()
	openaiCfg := buildDistillAIProviderConfig(cfgRepo.OpenAIProviderConfig(), distillProviderCfg)
	openaiRepo := openai.NewOpenAIRepository(openaiCfg, nil)

	containerBuilder := containerdi.New(openaiRepo, distillProviderCfg, "context-distill", serviceCfg)
	container, err := containerBuilder.Build()
	if err != nil {
		log.Fatal(err)
	}

	runner := (*container).Get(commands.RootCommandLabel).(commands.Runner)
	if err := runner.Execute(); err != nil {
		log.Fatal(err)
	}
}
