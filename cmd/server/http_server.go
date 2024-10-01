package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/routes"
	"github.com/gemyago/golang-backend-boilerplate/pkg/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/pkg/di"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

type runHTTPServerParams struct {
	dig.In `ignore-unexported:"true"`

	RootLogger *slog.Logger

	HTTPServer *http.Server

	ShutdownHandlers []di.ProcessShutdownHandler `group:"shutdown-handlers"`

	noop bool
}

func runHTTPServer(params runHTTPServerParams) error {
	rootLogger := params.RootLogger
	httpServer := params.HTTPServer
	rootCtx := context.Background()

	shutdown := func() {
		rootLogger.InfoContext(rootCtx, "Trying to shut down gracefully")
		ts := time.Now()

		grp := errgroup.Group{}
		for _, h := range params.ShutdownHandlers {
			grp.Go(func() error {
				rootLogger.InfoContext(rootCtx, fmt.Sprintf("Shutting down %s", h.Name))
				return h.Shutdown(rootCtx)
			})
		}

		// Not much we can do at this stage, so just logging
		if err := grp.Wait(); err != nil {
			rootLogger.ErrorContext(rootCtx, "Graceful shutdown failed", "err", err)
		}

		rootLogger.InfoContext(rootCtx, "Service stopped",
			slog.Duration("duration", time.Since(ts)),
		)
	}

	signalCtx, cancel := signal.NotifyContext(rootCtx, unix.SIGINT, unix.SIGTERM)
	defer cancel()

	startupErrors := make(chan error, 1)
	go func() {
		rootLogger.InfoContext(signalCtx, "Starting http listener",
			slog.String("addr", httpServer.Addr),
			slog.String("idleTimeout", httpServer.IdleTimeout.String()),
			slog.String("readHeaderTimeout", httpServer.ReadHeaderTimeout.String()),
			slog.String("readTimeout", httpServer.ReadTimeout.String()),
			slog.String("writeTimeout", httpServer.WriteTimeout.String()),
		)
		if params.noop {
			rootLogger.InfoContext(signalCtx, "NOOP: Exiting now")
			startupErrors <- nil
			return
		}
		startupErrors <- httpServer.ListenAndServe()
	}()

	var startupErr error
	select {
	case startupErr = <-startupErrors:
		if startupErr != nil {
			rootLogger.ErrorContext(rootCtx, "Server error", "err", startupErr)
		} else {
			rootLogger.InfoContext(rootCtx, "Server stopped")
		}
		shutdown()
	case <-signalCtx.Done(): // coverage-ignore
		shutdown()
	}
	return startupErr
}

func newHTTPServerCmd(container *dig.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "Command to start http server",
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
			// http related dependencies
			routes.Register(container),
			di.ProvideAll(
				container,
				server.NewHTTPServer,
				server.NewRootHandler,
			),
		)
	}
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return container.Invoke(func(params runHTTPServerParams) error {
			params.noop = noop
			return runHTTPServer(params)
		})
	}
	return cmd
}
