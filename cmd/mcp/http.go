package main

import (
	"context"

	mcpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type httpServerParams struct {
	dig.In `ignore-unexported:"true"`

	MCPServer *mcpserver.MCPServer
}

func startHTTPServer(params httpServerParams) error {
	rootCtx := context.Background()

	return params.MCPServer.StartHTTP(rootCtx)
}

func newHTTPCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Start MCP server with HTTP transport",
		Long:  "Start MCP server using HTTP transport for web-based MCP clients",
		RunE: func(cmd *cobra.Command, args []string) error {
			var params httpServerParams
			if err := container.Invoke(func(p httpServerParams) {
				params = p
			}); err != nil {
				return err
			}
			return startHTTPServer(params)
		},
	}

	cmd.Flags().StringP("host", "H", "localhost", "Host to bind HTTP server to")
	cmd.Flags().IntP("port", "p", 8080, "Port to bind HTTP server to")

	return cmd
}
