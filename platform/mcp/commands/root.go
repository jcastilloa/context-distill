package commands

import (
	"context"
	"fmt"
	"io"

	distillapp "github.com/jcastilloa/context-distill/distill/application/distillation"
	mcpserver "github.com/jcastilloa/context-distill/platform/mcp/server"
	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const RootCommandLabel = "mcp.root.command"

type Runner struct {
	serviceName         string
	serviceCfg          configDomain.ServiceConfig
	server              *mcpserver.Server
	configUIRunner      ConfigUIRunner
	distillBatchTool    tools.DistillBatch
	distillWatchTool    tools.DistillWatch
	distillBatchUseCase DistillBatchCLIUseCase
	distillWatchUseCase DistillWatchCLIUseCase
}

type DistillBatchCLIUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillBatchRequest) (distillapp.DistillBatchResult, error)
}

type DistillWatchCLIUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillWatchRequest) (distillapp.DistillWatchResult, error)
}

func NewRunner(
	serviceName string,
	serviceCfg configDomain.ServiceConfig,
	server *mcpserver.Server,
	configUIRunner ConfigUIRunner,
	distillBatchTool tools.DistillBatch,
	distillWatchTool tools.DistillWatch,
	distillBatchUseCase DistillBatchCLIUseCase,
	distillWatchUseCase DistillWatchCLIUseCase,
) Runner {
	return Runner{
		serviceName:         serviceName,
		serviceCfg:          serviceCfg,
		server:              server,
		configUIRunner:      configUIRunner,
		distillBatchTool:    distillBatchTool,
		distillWatchTool:    distillWatchTool,
		distillBatchUseCase: distillBatchUseCase,
		distillWatchUseCase: distillWatchUseCase,
	}
}

func (r Runner) Execute() error {
	return r.newRootCommand().Execute()
}

func (r Runner) newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   r.serviceName,
		Short: "MCP scaffold with Cobra + Viper + DI",
		RunE:  r.runServer(),
	}

	defaultTransport := r.serviceCfg.NormalizedTransport()
	cmd.Flags().String("transport", defaultTransport, "MCP transport (supported: stdio)")
	cmd.Flags().Bool("config-ui", false, "Open interactive terminal UI to configure distill provider settings")

	_ = viper.BindPFlag("service.transport", cmd.Flags().Lookup("transport"))
	viper.SetDefault("service.transport", defaultTransport)
	viper.SetEnvPrefix("MCP")
	viper.AutomaticEnv()

	cmd.AddCommand(r.newVersionCommand())
	cmd.AddCommand(r.newDistillBatchCommand())
	cmd.AddCommand(r.newDistillWatchCommand())
	return cmd
}

func (r Runner) runServer() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		configUIEnabled, _ := cmd.Flags().GetBool("config-ui")
		if configUIEnabled {
			return r.configUIRunner.Run(r.serviceName)
		}

		r.server.AddTool(r.distillBatchTool.Definition(), r.distillBatchTool.Handler)
		r.server.AddTool(r.distillWatchTool.Definition(), r.distillWatchTool.Handler)

		transport := viper.GetString("service.transport")
		if transport == "" {
			transport = r.serviceCfg.NormalizedTransport()
		}

		return r.server.Run(transport)
	}
}

func (r Runner) newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print service version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(r.serviceCfg.Version)
		},
	}
}

func (r Runner) newDistillBatchCommand() *cobra.Command {
	var question string
	var input string

	cmd := &cobra.Command{
		Use:   "distill_batch",
		Short: "Distill command output for one question",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := r.distillBatchUseCase.Execute(cmd.Context(), distillapp.DistillBatchRequest{
				Question: question,
				Input:    input,
			})
			if err != nil {
				return err
			}

			_, err = io.WriteString(cmd.OutOrStdout(), result.Output)
			return err
		},
	}

	cmd.Flags().StringVar(&question, "question", "", "Exact question to answer from the command output")
	cmd.Flags().StringVar(&input, "input", "", "Raw command output to distill")
	_ = cmd.MarkFlagRequired("question")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}

func (r Runner) newDistillWatchCommand() *cobra.Command {
	var question string
	var previousCycle string
	var currentCycle string

	cmd := &cobra.Command{
		Use:   "distill_watch",
		Short: "Distill changes between two watch snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := r.distillWatchUseCase.Execute(cmd.Context(), distillapp.DistillWatchRequest{
				Question:      question,
				PreviousCycle: previousCycle,
				CurrentCycle:  currentCycle,
			})
			if err != nil {
				return err
			}

			_, err = io.WriteString(cmd.OutOrStdout(), result.Output)
			return err
		},
	}

	cmd.Flags().StringVar(&question, "question", "", "Exact question to answer from cycle changes")
	cmd.Flags().StringVar(&previousCycle, "previous-cycle", "", "Previous watch cycle output")
	cmd.Flags().StringVar(&currentCycle, "current-cycle", "", "Current watch cycle output")
	_ = cmd.MarkFlagRequired("question")
	_ = cmd.MarkFlagRequired("previous-cycle")
	_ = cmd.MarkFlagRequired("current-cycle")

	return cmd
}
