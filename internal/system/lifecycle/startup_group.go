package lifecycle

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

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

// Start runs all startup functions and waits for shutdown signal.
// It also sets up a listener for forceful shutdown signals to
// terminate the application immediately if needed.
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

	shutdownSignals := []os.Signal{unix.SIGINT, unix.SIGTERM}

	shutdown := func() error {
		watchForceSignal(ctx, shutdownSignals)
		return g.shutdownHooks.PerformShutdown(ctx)
	}

	signalCtx, cancel := signal.NotifyContext(ctx, shutdownSignals...)
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
