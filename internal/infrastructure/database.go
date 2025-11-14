package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/dig"
)

type Database struct {
	instance *sql.DB
}

type DatabaseConfig struct {
	dig.In

	DSN string `name:"config.database.dsn"`
}

func newDBProvider(ctx context.Context) func(cfg DatabaseConfig) (*Database, error) {
	return func(cfg DatabaseConfig) (*Database, error) {
		db, err := sql.Open("sqlite", cfg.DSN)
		if err != nil { // coverage-ignore -- No way to simulate this
			return nil, fmt.Errorf("failed to open database: %w", err)
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
