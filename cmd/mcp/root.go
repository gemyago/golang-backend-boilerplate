package main

import (
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/controllers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/config"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

func newRootCmd(container *dig.Container) *cobra.Command {
	logsOutputFile := ""

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP (Model Context Protocol) server command",
		Long:  "Start MCP server with stdio or HTTP transport for providing tools to MCP clients",
	}
	cmd.SilenceUsage = true
	cmd.PersistentFlags().StringP("log-level", "l", "", "Produce logs with given level. Default is env specific.")
	cmd.PersistentFlags().StringVar(
		&logsOutputFile,
		"logs-file",
		"",
		"Produce logs to file instead of stdout. Used for tests only.",
	)
	cmd.PersistentFlags().Bool(
		"json-logs",
		false,
		"Indicates if logs should be in JSON format or text (default)",
	)
	cmd.PersistentFlags().StringP(
		"env",
		"e",
		"",
		"Env that the process is running in.",
	)
	cfg := config.New()
	lo.Must0(cfg.BindPFlags(cmd.PersistentFlags()))
	cmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		rootCtx := cmd.Context()

		return errors.Join(
			internal.Setup(
				rootCtx,
				cfg,
				container,
			),

			// mcp components
			controllers.Register(container),

			di.ProvideAll(container,
				server.NewMCPServer,
			),
		)
	}
	return cmd
}
