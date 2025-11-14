package app

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/dig"
)

// UserQueries is a concrete struct (not an interface).
// Controllers and other consumers use this directly.
// Follows "accept interface, return struct" principle.
type UserQueries struct {
	usersRepo UsersRepository
	logger    *slog.Logger
}

type UserQueriesDeps struct {
	dig.In

	UsersRepo  UsersRepository
	RootLogger *slog.Logger
}

// NewUserQueries returns a concrete struct (not an interface).
// This follows "accept interface, return struct" principle.
func NewUserQueries(deps UserQueriesDeps) *UserQueries {
	return &UserQueries{
		usersRepo: deps.UsersRepo,
		logger:    deps.RootLogger.WithGroup("app.user-queries"),
	}
}

func (q *UserQueries) GetUserByID(_ context.Context, _ string) (*User, error) {
	// Fetch from repository.
	// Return user (without timestamps for API response).
	return nil, errors.New("not implemented")
}

func (q *UserQueries) ListUsers(_ context.Context) ([]*User, error) {
	// Fetch all from repository.
	return nil, errors.New("not implemented")
}
