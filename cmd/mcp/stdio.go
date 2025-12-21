package main

import (
	"context"
	"log/slog"
	"os"

	mcpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type stdioServerParams struct {
	dig.In `ignore-unexported:"true"`

	StartupGroupFactory lifecycle.StartupGroupFactory

	RootLogger *slog.Logger

	MCPServer *mcpserver.MCPServer

	noop bool
}

func startStdioServer(rootCtx context.Context, params stdioServerParams) error {
	rootLogger := params.RootLogger

	startupGroup := params.StartupGroupFactory.NewGroup()
	startupGroup.Add(func(ctx context.Context) error {
		if params.noop {
			rootLogger.InfoContext(ctx, "NOOP: Starting stdio server")
			return nil
		}
		return params.MCPServer.ListenStdioServer(ctx, os.Stdin, os.Stdout)
	})
	return startupGroup.Start(rootCtx)
}

func newStdioCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stdio",
		Short: "Start MCP server with stdio transport",
		Long:  "Start MCP server using stdio transport for communication with MCP clients",
	}

	noop := false
	cmd.Flags().BoolVar(
		&noop,
		"noop",
		false,
		"Run in noop mode. Useful for testing if setup is all working.",
	)
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(p stdioServerParams) error {
			p.noop = noop
			return startStdioServer(cmd.Context(), p)
		})
	}

	return cmd
}
