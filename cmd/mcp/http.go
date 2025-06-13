package main

import (
	mcpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

const defaultHTTPPort = 8080

type httpServerParams struct {
	dig.In `ignore-unexported:"true"`

	MCPServer *mcpserver.MCPServer
}

func newHTTPCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Start MCP server with HTTP transport",
		Long:  "Start MCP server using HTTP transport for web-based MCP clients",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return container.Invoke(func(p httpServerParams) error {
				return p.MCPServer.StartHTTP(cmd.Context())
			})
		},
	}

	cmd.Flags().StringP("host", "H", "localhost", "Host to bind HTTP server to")
	cmd.Flags().IntP("port", "p", defaultHTTPPort, "Port to bind HTTP server to")

	return cmd
}
