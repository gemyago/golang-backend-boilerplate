//go:build !release

package infrastructure

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/config"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type TestDatabaseOpts func(deps *DatabaseDeps)

func NewTestDatabase(t *testing.T, opts ...TestDatabaseOpts) *Database {
	t.Helper()

	cfg := config.New()

	env := cfg.GetString("env")
	if env == "" {
		env = "test"
	}

	err := config.Load(cfg, config.NewLoadOpts().WithEnv(env))
	require.NoError(t, err)

	deps := DatabaseDeps{
		ShutdownHooks:  lifecycle.NewTestShutdownHooks(),
		MeterProvider:  metricnoop.NewMeterProvider(),
		TracerProvider: tracenoop.NewTracerProvider(),
		DSN:            cfg.GetString("database.dsn"),
	}
	for _, opt := range opts {
		opt(&deps)
	}
	db, err := newDBProvider(t.Context())(deps)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.instance.Close()
	})

	return db
}
