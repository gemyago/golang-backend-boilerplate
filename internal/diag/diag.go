package diag

import "context"

type ShutdownHooks interface {
	Register(name string, hook func(ctx context.Context) error)
}
