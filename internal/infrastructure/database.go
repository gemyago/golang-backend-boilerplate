package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type Database struct {
	instance *sql.DB
}

func (db *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.instance.QueryContext(ctx, query, args...)
}

func (db *Database) Close() error {
	return db.instance.Close()
}

type DatabaseDeps struct {
	dig.In

	metric.MeterProvider
	trace.TracerProvider
	*lifecycle.ShutdownHooks

	OTELConfig telemetry.OTELConfig

	DSN string `name:"config.database.dsn"`
}

func newDBProvider(ctx context.Context) func(DatabaseDeps) (*Database, error) {
	return func(deps DatabaseDeps) (*Database, error) {
		var db *sql.DB
		var err error

		const driverName = "sqlite"

		if deps.OTELConfig.Enabled {
			db, err = otelsql.Open(driverName, deps.DSN,
				otelsql.WithAttributes(semconv.DBSystemSqlite),
				otelsql.WithMeterProvider(deps.MeterProvider),
				otelsql.WithTracerProvider(deps.TracerProvider),
			)
		} else {
			db, err = sql.Open(driverName, deps.DSN)
		}

		if err != nil { // coverage-ignore -- No way to simulate this
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		// Enable foreign key enforcement so ON DELETE CASCADE works as expected.
		// By default SQLite requires PRAGMA foreign_keys = ON per connection.
		// The error path below is hard to simulate in tests (driver-level failures),
		// so exclude it from coverage measurements.
		if _, err = db.ExecContext(
			ctx,
			"PRAGMA foreign_keys = ON;",
		); err != nil { // coverage-ignore -- No way to simulate this
			// Attempt to close DB and report both errors if close fails.
			if closeErr := db.Close(); closeErr != nil {
				// Use %s for the secondary error and wrap the primary using %w.
				return nil, fmt.Errorf(
					"failed to enable foreign keys: %w; additionally failed to close db: %s",
					err,
					closeErr.Error(),
				)
			}
			return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
		}

		// This is minimalistic setup not intended for production use.
		// In a real-world application, you would handle migrations, connection pooling, etc.
		// Migrations are very likely to be handled outside by a dedicated tool.

		if err = errors.Join(
			// Add other schema entries as needed
			initUsersSchema(ctx, db),
			initUserPetsSchema(ctx, db),
		); err != nil {
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}

		deps.ShutdownHooks.RegisterNoCtx("database", db.Close)

		return &Database{instance: db}, nil
	}
}

func initUsersSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func initUserPetsSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS user_pets (
			user_id TEXT NOT NULL,
			pet_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, pet_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`)
	return err
}
