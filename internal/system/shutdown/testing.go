//go:build !release

package shutdown

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
)

const defaultTestShutdownTimeout = 30 * time.Second

// NewTestHooks constructor for shutdown Hooks that can be used in tests.
func NewTestHooks() *Hooks {
	return NewHooks(HooksDeps{
		RootLogger:              diag.RootTestLogger(),
		GracefulShutdownTimeout: defaultTestShutdownTimeout,
	})
}
