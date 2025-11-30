package app

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type UsersMetrics struct {
	usersCreated       metric.Int64Counter
	usersEmailConflict metric.Int64Counter
}

func newUsersMetrics(
	provider metric.MeterProvider,
) (*UsersMetrics, error) { // coverage-ignore -- coverage drops on error handling, hard to simulate this in tests
	meter := provider.Meter("app.metrics")

	usersCreated, err := meter.Int64Counter("users.created")

	if err != nil {
		return nil, fmt.Errorf("failed to create users.created counter: %w", err)
	}

	usersEmailConflict, err := meter.Int64Counter("users.email_conflict")

	if err != nil {
		return nil, fmt.Errorf("failed to create users.email_conflict counter: %w", err)
	}

	return &UsersMetrics{
		usersCreated:       usersCreated,
		usersEmailConflict: usersEmailConflict,
	}, nil
}

func (m *UsersMetrics) recordUserCreated(ctx context.Context) {
	m.usersCreated.Add(ctx, 1)
}

func (m *UsersMetrics) recordUserEmailConflict(ctx context.Context) {
	m.usersEmailConflict.Add(ctx, 1)
}
