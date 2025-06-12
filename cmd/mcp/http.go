package main

import (
	"context"
	"log/slog"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type httpServerParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger
	// TODO: Add MCP server dependencies in later tasks
}

func startHTTPServer(params httpServerParams) error {
	rootLogger := params.RootLogger
	rootCtx := context.Background()

	rootLogger.InfoContext(rootCtx, "Starting MCP server with HTTP transport")

	// TODO: Implement actual MCP HTTP server in task 2.0
	rootLogger.InfoContext(rootCtx, "MCP HTTP server would start here")

	return nil
}

func newHTTPCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Start MCP server with HTTP transport",
		Long:  "Start MCP server using HTTP transport for communication with MCP clients",
	}

	// Add HTTP-specific flags
	var port int
	var host string

	cmd.Flags().IntVar(&port, "port", 8080, "Port to listen on for HTTP transport")
	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind to for HTTP transport")

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(params httpServerParams) error {
			params.RootLogger.InfoContext(context.Background(),
				"HTTP server configuration",
				slog.String("host", host),
				slog.Int("port", port))
			return startHTTPServer(params)
		})
	}

	return cmd
}
