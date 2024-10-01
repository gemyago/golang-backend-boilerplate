package di

import (
	"context"
)

// ProcessShutdownHandler is used mostly to register
// shutdown handlers that should perform cleanup tasks on
// process shutdown.
type ProcessShutdownHandler struct {
	// Name is used for logging purposes
	Name string

	// Shutdown is the function that will perform the cleanup
	Shutdown func(ctx context.Context) error
}

func MakeProcessShutdownHandler(name string, fn func(ctx context.Context) error) ProcessShutdownHandler {
	return ProcessShutdownHandler{name, fn}
}

func MakeProcessShutdownHandlerNoContext(name string, fn func() error) ProcessShutdownHandler {
	return ProcessShutdownHandler{name, func(context.Context) error {
		return fn()
	}}
}
