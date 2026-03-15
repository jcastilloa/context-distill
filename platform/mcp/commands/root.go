package commands

import (
	"fmt"

	mcpserver "github.com/jcastilloa/context-distill/platform/mcp/server"
	"github.com/jcastilloa/context-distill/platform/mcp/tools"
	configDomain "github.com/jcastilloa/context-distill/shared/config/domain"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const RootCommandLabel = "mcp.root.command"

type Runner struct {
	serviceName      string
	serviceCfg       configDomain.ServiceConfig
	server           *mcpserver.Server
	configUIRunner   ConfigUIRunner
	distillBatchTool tools.DistillBatch
	distillWatchTool tools.DistillWatch
}

func NewRunner(
	serviceName string,
	serviceCfg configDomain.ServiceConfig,
	server *mcpserver.Server,
	configUIRunner ConfigUIRunner,
	distillBatchTool tools.DistillBatch,
	distillWatchTool tools.DistillWatch,
) Runner {
	return Runner{
		serviceName:      serviceName,
		serviceCfg:       serviceCfg,
		server:           server,
		configUIRunner:   configUIRunner,
		distillBatchTool: distillBatchTool,
		distillWatchTool: distillWatchTool,
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
