package diag

import "context"

const CorrelationIDHeader = "X-Correlation-ID"

type ShutdownHooks interface {
	Register(name string, hook func(ctx context.Context) error)
}
