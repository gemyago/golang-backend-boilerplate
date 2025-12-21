package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

type startServerParams struct {
	dig.In `ignore-unexported:"true"`

	StartupGroupFactory lifecycle.StartupGroupFactory
	RootLogger          *slog.Logger

	HTTPServer *server.HTTPServer

	noop bool
}

func startServer(params startServerParams) error {
	rootLogger := params.RootLogger
	httpServer := params.HTTPServer
	rootCtx := context.Background()

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

func newStartServerCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Command to start server",
	}
	noop := false
	cmd.Flags().BoolVar(
		&noop,
		"noop",
		false,
		"Do not start. Just setup deps and exit. Useful for testing if setup is all working.",
	)
	cmd.PreRunE = func(_ *cobra.Command, _ []string) error {
		return errors.Join(
			server.Register(container),
			http.Register(container),
		)
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(params startServerParams) error {
			params.noop = noop
			return startServer(params)
		})
	}
	return cmd
}
