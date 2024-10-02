// Components in this packages are used to implement a graceful shutdown
// of the application. This may include closing database connections, flushing pending
// events to the queue, shutting down the http server, etc.

package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

type ShutdownHook interface {
	// Name returns the name of the shutdown hook
	// for logging purposes
	Name() string

	// Shutdown is the function that will perform the cleanup
	// on shutdown of the process
	Shutdown(ctx context.Context) error
}

type shutdownHook struct {
	name     string
	shutdown func(ctx context.Context) error
}

func (s *shutdownHook) Name() string {
	return s.name
}

func (s *shutdownHook) Shutdown(ctx context.Context) error {
	return s.shutdown(ctx)
}

func NewShutdownHookNoCtx(name string, shutdown func() error) ShutdownHook {
	return &shutdownHook{
		name: name,
		shutdown: func(_ context.Context) error {
			return shutdown()
		},
	}
}

type ShutdownHooks interface {
	Register(hook ShutdownHook)
	PerformShutdown(ctx context.Context) error
}

type shutdownHooks struct {
	logger *slog.Logger
	hooks  []ShutdownHook
	ShutdownHooksRegistryDeps
}

func (s *shutdownHooks) Register(hook ShutdownHook) {
	s.hooks = append(s.hooks, hook)
}

func (s *shutdownHooks) PerformShutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.GracefulShutdownTimeout)
	defer cancel()

	errGrp := errgroup.Group{}
	for _, hook := range s.hooks {
		errGrp.Go(func() error {
			hookName := hook.Name()
			s.logger.InfoContext(ctx, fmt.Sprintf("Shutting down %s", hookName))
			if err := hook.Shutdown(ctx); err != nil {
				return fmt.Errorf("failed to perform shutdown hook %s: %w", hookName, err)
			}
			return nil
		})
	}

	done := make(chan error)
	go func() {
		done <- errGrp.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type ShutdownHooksRegistryDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	GracefulShutdownTimeout time.Duration `name:"config.gracefulShutdownTimeout"`
}

func NewShutdownHooks(deps ShutdownHooksRegistryDeps) ShutdownHooks {
	return &shutdownHooks{
		logger:                    deps.RootLogger.WithGroup("shutdown"),
		ShutdownHooksRegistryDeps: deps,
	}
}
