//go:build !release

package lifecycle

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

const defaultTestShutdownTimeout = 30 * time.Second

// NewTestShutdownHooks constructor for ShutdownHooks that can be used in tests.
func NewTestShutdownHooks() *Hooks {
	return NewHooks(HooksDeps{
		RootLogger:              diag.RootTestLogger(),
		GracefulShutdownTimeout: defaultTestShutdownTimeout,
	})
}
