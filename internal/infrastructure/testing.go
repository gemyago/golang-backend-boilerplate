//go:build !release

package infrastructure

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

const defaultTestShutdownTimeout = 30 * time.Second

func NewTestShutdownHooks() *ShutdownHooks {
	return NewShutdownHooks(ShutdownHooksDeps{
		RootLogger:              diag.RootTestLogger(),
		GracefulShutdownTimeout: defaultTestShutdownTimeout,
	})
}
