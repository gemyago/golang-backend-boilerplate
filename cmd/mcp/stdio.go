package main

import (
	"context"

	mcpserver "github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type stdioServerParams struct {
	dig.In `ignore-unexported:"true"`

	MCPServer *mcpserver.MCPServer
}

func startStdioServer(params stdioServerParams) error {
	rootCtx := context.Background()

	return params.MCPServer.StartStdio(rootCtx)
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
