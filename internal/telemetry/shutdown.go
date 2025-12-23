package telemetry

import (
	"context"
	"log/slog"
)

type ShutdownHooks interface {
	Register(name string, hook func(ctx context.Context) error)
}

type shutdowner interface {
	Shutdown(ctx context.Context) error
}

func registerShutdownHook(logger *slog.Logger, hooks ShutdownHooks, name string, target any) { // coverage-ignore
	switch target := target.(type) {
	case shutdowner:
		hooks.Register(name, target.Shutdown)
	default:
		logger.Warn("registerShutdownHook: target is not a shutdowner",
			slog.Any("target", target),
			slog.String("name", name),
		)
		return
	}
}
