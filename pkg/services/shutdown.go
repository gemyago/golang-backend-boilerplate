package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.uber.org/dig"
)

type ShutdownHook interface {
	// Name returns the name of the shutdown hook
	// for logging purposes
	Name() string

	// Shutdown is the function that will perform the cleanup
	// on shutdown of the process
	Shutdown(ctx context.Context) error
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
	for _, hook := range s.hooks {
		hookName := hook.Name()
		s.logger.InfoContext(ctx, "Performing shutdown hook", slog.String("hook", hookName))
		if err := hook.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to perform shutdown hook %s: %w", hookName, err)
		}
	}
	return nil
}

type ShutdownHooksRegistryDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	MaxShutdownDuration time.Duration `name:"config.shutdown.maxDuration"`
}

func NewShutdownHooksRegistry(deps ShutdownHooksRegistryDeps) ShutdownHooks {
	return &shutdownHooks{
		logger: deps.RootLogger.WithGroup("shutdown"),
	}
}
