package lifecycle

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"go.uber.org/dig"
	"golang.org/x/sys/unix"
)

type StartupGroupFactory struct {
	dig.In

	ShutdownHooks *ShutdownHooks
	RootLogger    *slog.Logger
}

func (f *StartupGroupFactory) NewGroup() *StartupGroup {
	return &StartupGroup{
		shutdownHooks: f.ShutdownHooks,
		logger:        f.RootLogger.WithGroup("startup-group"),
	}
}

type StartupFn func(ctx context.Context) error

type StartupGroup struct {
	startupFns []StartupFn

	shutdownHooks *ShutdownHooks
	logger        *slog.Logger
}

func (g *StartupGroup) Add(fn StartupFn) {
	g.startupFns = append(g.startupFns, fn)
}

func (g *StartupGroup) Start(ctx context.Context) error {
	watchForceSignal := func(
		rootCtx context.Context,
		signals []os.Signal,
	) {
		forceSignal := make(chan os.Signal, 1)
		signal.Notify(forceSignal, signals...)

		go func() {
			<-forceSignal
			g.logger.InfoContext(rootCtx, "Forcing shutdown")
			os.Exit(1)
		}()
	}

	shutdown := func() error {
		watchForceSignal(ctx, []os.Signal{unix.SIGINT, unix.SIGTERM})

		g.logger.InfoContext(ctx, "Attempting to shut down gracefully")
		ts := time.Now()

		err := g.shutdownHooks.PerformShutdown(ctx)
		if err != nil {
			g.logger.ErrorContext(ctx, "Failed to shut down gracefully", telemetry.ErrAttr(err))
		}

		g.logger.InfoContext(ctx, "Application stopped",
			slog.Duration("duration", time.Since(ts)),
		)
		return err
	}

	signalCtx, cancel := signal.NotifyContext(ctx, unix.SIGINT, unix.SIGTERM)
	defer cancel()

	startupErrors := make(chan error, len(g.startupFns))
	for _, fn := range g.startupFns {
		go func(fn StartupFn) {
			startupErrors <- fn(signalCtx)
		}(fn)
	}

	var startupErr error
	select {
	case startupErr = <-startupErrors:
		if startupErr != nil {
			g.logger.ErrorContext(ctx, "Application startup failed", telemetry.ErrAttr(startupErr))
		}
	case <-signalCtx.Done(): // coverage-ignore
		// We will attempt to shut down in both cases
		// so doing it once on a next line
	}
	return errors.Join(startupErr, shutdown())
}
