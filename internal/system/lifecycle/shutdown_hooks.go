// Hooks are used to perform a graceful shutdown
// of the application. This may include closing database connections,
// shutting down the http server, etc.

package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"go.uber.org/dig"
)

type shutdownHook struct {
	name       string
	shutdownFn func(ctx context.Context) error
}

type ShutdownHooksDeps struct {
	dig.In

	RootLogger *slog.Logger

	// config
	GracefulShutdownTimeout time.Duration `name:"config.gracefulShutdownTimeout"`
}

type ShutdownHooks struct {
	logger *slog.Logger
	hooks  []shutdownHook
	deps   ShutdownHooksDeps
}

// NewShutdownHooks creates a new instance of Hooks.
func NewShutdownHooks(deps ShutdownHooksDeps) *ShutdownHooks {
	return &ShutdownHooks{
		logger: deps.RootLogger.WithGroup("shutdown"),
		deps:   deps,
	}
}

// HasHook checks if a shutdown hook with the given name is registered.
// Typical usage is in tests and must be carefully considered for production scenarios.
func (h *ShutdownHooks) HasHook(name string, method any) bool {
	for _, hook := range h.hooks {
		if hook.name == name {
			return reflect.ValueOf(hook.shutdownFn).Pointer() == reflect.ValueOf(method).Pointer()
		}
	}
	return false
}

func (h *ShutdownHooks) Register(name string, shutdown func(ctx context.Context) error) {
	h.hooks = append(h.hooks, shutdownHook{name: name, shutdownFn: shutdown})
}

func (h *ShutdownHooks) RegisterNoCtx(name string, shutdown func() error) {
	h.Register(name, func(_ context.Context) error {
		return shutdown()
	})
}

func (h *ShutdownHooks) PerformShutdown(ctx context.Context) error {
	ts := time.Now()
	defer func() {
		h.logger.InfoContext(ctx, "Application stopped",
			slog.Duration("duration", time.Since(ts)),
		)
	}()
	h.logger.InfoContext(ctx, "Attempting to shut down gracefully")

	ctx, cancel := context.WithTimeout(ctx, h.deps.GracefulShutdownTimeout)
	defer cancel()

	resultsChan := make(chan error, len(h.hooks))

	var wg sync.WaitGroup

	for _, hook := range h.hooks {
		wg.Go(func() {
			hookName := hook.name
			h.logger.InfoContext(ctx, fmt.Sprintf("Shutting down %s", hookName))
			if err := hook.shutdownFn(ctx); err != nil {
				resultsChan <- fmt.Errorf("failed to perform shutdown hook %s: %w", hookName, err)
			} else {
				resultsChan <- nil
			}
		})
	}

	done := make(chan error)
	go func() {
		wg.Wait()
		close(resultsChan)
		errs := make([]error, 0)
		for err := range resultsChan {
			if err != nil {
				errs = append(errs, err)
			}
		}

		done <- errors.Join(errs...)
	}()

	select {
	case err := <-done:
		if err != nil {
			h.logger.ErrorContext(ctx, "Failed to shut down gracefully", telemetry.ErrAttr(err))
		}
		return err
	case <-ctx.Done():
		h.logger.WarnContext(ctx, "Shutdown hooks did not complete timely", telemetry.ErrAttr(ctx.Err()))
		return fmt.Errorf("shutdown hooks did not complete timely: %w", ctx.Err())
	}
}
