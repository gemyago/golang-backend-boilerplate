package main

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type stdioServerParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger
	// TODO: Add MCP server dependencies in later tasks
}

func startStdioServer(params stdioServerParams) error {
	rootLogger := params.RootLogger
	rootCtx := context.Background()

	rootLogger.InfoContext(rootCtx, "Starting MCP server with stdio transport")

	// TODO: Implement actual MCP stdio server in task 2.0
	rootLogger.InfoContext(rootCtx, "MCP stdio server would start here")

	return nil
}

func newStdioCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stdio",
		Short: "Start MCP server with stdio transport",
		Long:  "Start MCP server using stdio transport for communication with MCP clients",
	}

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(params stdioServerParams) error {
			return startStdioServer(params)
		})
	}

	return cmd
}
