package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

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
	searchCodeTool      tools.SearchCode
	distillBatchUseCase DistillBatchCLIUseCase
	distillWatchUseCase DistillWatchCLIUseCase
	searchCodeUseCase   SearchCodeCLIUseCase
}

type DistillBatchCLIUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillBatchRequest) (distillapp.DistillBatchResult, error)
}

type DistillWatchCLIUseCase interface {
	Execute(ctx context.Context, request distillapp.DistillWatchRequest) (distillapp.DistillWatchResult, error)
}

type SearchCodeCLIUseCase interface {
	Execute(ctx context.Context, request distillapp.SearchCodeRequest) (distillapp.SearchCodeResult, error)
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

func (r Runner) WithSearchCode(searchCodeTool tools.SearchCode, searchCodeUseCase SearchCodeCLIUseCase) Runner {
	r.searchCodeTool = searchCodeTool
	r.searchCodeUseCase = searchCodeUseCase
	return r
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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	cmd.AddCommand(r.newVersionCommand())
	cmd.AddCommand(r.newDistillBatchCommand())
	cmd.AddCommand(r.newDistillWatchCommand())
	cmd.AddCommand(r.newSearchCodeCommand())
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
		if r.searchCodeUseCase != nil {
			r.server.AddTool(r.searchCodeTool.Definition(), r.searchCodeTool.Handler)
		}

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
			resolvedInput, err := resolveInput(input, cmd.InOrStdin())
			if err != nil {
				return err
			}

			result, err := r.distillBatchUseCase.Execute(cmd.Context(), distillapp.DistillBatchRequest{
				Question: question,
				Input:    resolvedInput,
			})
			if err != nil {
				return err
			}

			_, err = io.WriteString(cmd.OutOrStdout(), result.Output)
			return err
		},
	}

	cmd.Flags().StringVar(&question, "question", "", "Exact question to answer from the command output")
	cmd.Flags().StringVar(&input, "input", "", "Raw command output to distill (reads from stdin if omitted or \"-\")")
	_ = cmd.MarkFlagRequired("question")

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

func (r Runner) newSearchCodeCommand() *cobra.Command {
	var query string
	var mode string
	var question string
	var scope []string
	var maxResults int
	var contextLines int

	cmd := &cobra.Command{
		Use:   "search_code",
		Short: "Search repository code and return compact distilled output",
		RunE: func(cmd *cobra.Command, args []string) error {
			if r.searchCodeUseCase == nil {
				return fmt.Errorf("search_code use case is not configured")
			}

			request := distillapp.SearchCodeRequest{
				Query:        resolveStringFlagValue(cmd, "query", "search_code.query", query),
				Mode:         resolveStringFlagValue(cmd, "mode", "search_code.mode", mode),
				Question:     resolveStringFlagValue(cmd, "question", "search_code.question", question),
				Scope:        resolveStringSliceFlagValue(cmd, "scope", "search_code.scope", scope),
				MaxResults:   resolveIntFlagValue(cmd, "max-results", "search_code.max_results", maxResults),
				ContextLines: resolveIntFlagValue(cmd, "context-lines", "search_code.context_lines", contextLines),
			}

			result, err := r.searchCodeUseCase.Execute(cmd.Context(), request)
			if err != nil {
				return err
			}

			_, err = io.WriteString(cmd.OutOrStdout(), result.Output)
			return err
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Search query")
	cmd.Flags().StringVar(&mode, "mode", "", "Search mode: text|regex|symbol|path")
	cmd.Flags().StringVar(&question, "question", "", "Output contract for distilled result")
	cmd.Flags().StringSliceVar(&scope, "scope", []string{}, "Optional scope globs (repeat flag or use comma-separated values)")
	cmd.Flags().IntVar(&maxResults, "max-results", distillapp.DefaultSearchCodeMaxResults, "Hard limit for returned candidates")
	cmd.Flags().IntVar(&contextLines, "context-lines", distillapp.DefaultSearchCodeContextLines, "Context lines per match")

	_ = viper.BindPFlag("search_code.query", cmd.Flags().Lookup("query"))
	_ = viper.BindPFlag("search_code.mode", cmd.Flags().Lookup("mode"))
	_ = viper.BindPFlag("search_code.question", cmd.Flags().Lookup("question"))
	_ = viper.BindPFlag("search_code.scope", cmd.Flags().Lookup("scope"))
	_ = viper.BindPFlag("search_code.max_results", cmd.Flags().Lookup("max-results"))
	_ = viper.BindPFlag("search_code.context_lines", cmd.Flags().Lookup("context-lines"))

	viper.SetDefault("search_code.query", "")
	viper.SetDefault("search_code.mode", "")
	viper.SetDefault("search_code.question", "")
	viper.SetDefault("search_code.scope", []string{})
	viper.SetDefault("search_code.max_results", distillapp.DefaultSearchCodeMaxResults)
	viper.SetDefault("search_code.context_lines", distillapp.DefaultSearchCodeContextLines)

	return cmd
}

func resolveStringFlagValue(cmd *cobra.Command, flagName, viperKey, currentValue string) string {
	if cmd.Flags().Changed(flagName) {
		return currentValue
	}
	return viper.GetString(viperKey)
}

func resolveStringSliceFlagValue(cmd *cobra.Command, flagName, viperKey string, currentValue []string) []string {
	if cmd.Flags().Changed(flagName) {
		return currentValue
	}
	return viper.GetStringSlice(viperKey)
}

func resolveIntFlagValue(cmd *cobra.Command, flagName, viperKey string, currentValue int) int {
	if cmd.Flags().Changed(flagName) {
		return currentValue
	}
	return viper.GetInt(viperKey)
}
