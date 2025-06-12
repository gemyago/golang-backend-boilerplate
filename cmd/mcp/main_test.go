package main

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

func TestSetupCommands(t *testing.T) {
	t.Run("should create root command with proper configuration", func(t *testing.T) {
		rootCmd := setupCommands()

		require.NotNil(t, rootCmd)
		require.Equal(t, "mcp", rootCmd.Use)
		require.Equal(t, "MCP (Model Context Protocol) server command", rootCmd.Short)
		require.Contains(t, rootCmd.Long, "Start MCP server with stdio or HTTP transport")
	})

	t.Run("should register stdio and http subcommands", func(t *testing.T) {
		rootCmd := setupCommands()

		commands := rootCmd.Commands()
		commandNames := make([]string, 0, len(commands))
		for _, cmd := range commands {
			commandNames = append(commandNames, cmd.Use)
		}

		require.Contains(t, commandNames, "stdio")
		require.Contains(t, commandNames, "http")
	})

	t.Run("should have persistent flags configured", func(t *testing.T) {
		rootCmd := setupCommands()

		// Check that persistent flags are present
		require.NotNil(t, rootCmd.PersistentFlags().Lookup("log-level"))
		require.NotNil(t, rootCmd.PersistentFlags().Lookup("logs-file"))
		require.NotNil(t, rootCmd.PersistentFlags().Lookup("json-logs"))
		require.NotNil(t, rootCmd.PersistentFlags().Lookup("env"))
	})
}

func TestNewRootCmd(t *testing.T) {
	t.Run("should create root command with correct usage", func(t *testing.T) {
		container := dig.New()
		cmd := newRootCmd(container)

		require.Equal(t, "mcp", cmd.Use)
		require.Equal(t, "MCP (Model Context Protocol) server command", cmd.Short)
		require.True(t, cmd.SilenceUsage)
	})

	t.Run("should have all required persistent flags", func(t *testing.T) {
		container := dig.New()
		cmd := newRootCmd(container)

		flags := []string{"log-level", "logs-file", "json-logs", "env"}
		for _, flagName := range flags {
			flag := cmd.PersistentFlags().Lookup(flagName)
			require.NotNil(t, flag, "Flag %s should be present", flagName)
		}
	})
}

func TestNewStdioCmd(t *testing.T) {
	t.Run("should create stdio command with correct configuration", func(t *testing.T) {
		container := dig.New()
		cmd := newStdioCmd(container)

		require.Equal(t, "stdio", cmd.Use)
		require.Equal(t, "Start MCP server with stdio transport", cmd.Short)
		require.Contains(t, cmd.Long, "stdio transport for communication")
		require.NotNil(t, cmd.RunE)
	})
}

func TestNewHTTPCmd(t *testing.T) {
	t.Run("should create http command with correct configuration", func(t *testing.T) {
		container := dig.New()
		cmd := newHTTPCmd(container)

		require.Equal(t, "http", cmd.Use)
		require.Equal(t, "Start MCP server with HTTP transport", cmd.Short)
		require.Contains(t, cmd.Long, "HTTP transport for communication")
		require.NotNil(t, cmd.RunE)
	})

	t.Run("should have port and host flags", func(t *testing.T) {
		container := dig.New()
		cmd := newHTTPCmd(container)

		portFlag := cmd.Flags().Lookup("port")
		require.NotNil(t, portFlag)
		require.Equal(t, "8080", portFlag.DefValue)

		hostFlag := cmd.Flags().Lookup("host")
		require.NotNil(t, hostFlag)
		require.Equal(t, "localhost", hostFlag.DefValue)
	})

	t.Run("should accept custom port and host values", func(t *testing.T) {
		container := dig.New()
		cmd := newHTTPCmd(container)

		// Use faker to generate test host value
		testHost := faker.IPv4()

		args := []string{"--port", "9090", "--host", testHost}
		cmd.SetArgs(args)

		err := cmd.Flags().Parse(args)
		require.NoError(t, err)

		port, err := cmd.Flags().GetInt("port")
		require.NoError(t, err)
		require.Equal(t, 9090, port)

		host, err := cmd.Flags().GetString("host")
		require.NoError(t, err)
		require.Equal(t, testHost, host)
	})
}

func TestCommandIntegration(t *testing.T) {
	t.Run("should have stdio subcommand available", func(t *testing.T) {
		rootCmd := setupCommands()

		// Find stdio command
		var stdioCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "stdio" {
				stdioCmd = cmd
				break
			}
		}

		require.NotNil(t, stdioCmd, "stdio command should be available")
		require.Equal(t, "stdio", stdioCmd.Use)
		require.Contains(t, stdioCmd.Short, "stdio transport")
	})

	t.Run("should have http subcommand available", func(t *testing.T) {
		rootCmd := setupCommands()

		// Find HTTP command
		var httpCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == "http" {
				httpCmd = cmd
				break
			}
		}

		require.NotNil(t, httpCmd, "http command should be available")
		require.Equal(t, "http", httpCmd.Use)
		require.Contains(t, httpCmd.Short, "HTTP transport")
		require.NotNil(t, httpCmd.Flags().Lookup("port"))
		require.NotNil(t, httpCmd.Flags().Lookup("host"))
	})

	t.Run("should register commands properly in root", func(t *testing.T) {
		rootCmd := setupCommands()

		commands := rootCmd.Commands()
		require.True(t, len(commands) >= 2, "Should have at least stdio and http commands")

		commandNames := make(map[string]bool)
		for _, cmd := range commands {
			commandNames[cmd.Use] = true
		}

		require.True(t, commandNames["stdio"], "stdio command should be registered")
		require.True(t, commandNames["http"], "http command should be registered")
	})
}
