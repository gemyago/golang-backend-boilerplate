package main

import (
	"context"
	"log/slog"

	mcpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type startHTTPServerParams struct {
	dig.In `ignore-unexported:"true"`

	StartupGroupFactory lifecycle.StartupGroupFactory

	RootLogger *slog.Logger

	MCPServer *mcpserver.MCPServer

	noop bool
}

func startHTTPServer(rootCtx context.Context, params startHTTPServerParams) error {
	rootLogger := params.RootLogger
	httpServer := params.MCPServer.NewStreamableHTTPServer()

	startupGroup := params.StartupGroupFactory.NewGroup()
	startupGroup.Add(func(ctx context.Context) error {
		if params.noop {
			rootLogger.InfoContext(ctx, "NOOP: Starting http server")
			return nil
		}
		return httpServer.Start(ctx)
	})

	return startupGroup.Start(rootCtx)
}

func newHTTPCmd(container *dig.Container) *cobra.Command {
	noop := false
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Start MCP server with HTTP transport",
		Long:  "Start MCP server using HTTP transport for web-based MCP clients",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return container.Invoke(func(p startHTTPServerParams) error {
				p.noop = noop
				return startHTTPServer(cmd.Context(), p)
			})
		},
	}
	cmd.Flags().BoolVar(
		&noop,
		"noop",
		false,
		"Run in noop mode. Useful for testing if setup is all working.",
	)

	return cmd
}
