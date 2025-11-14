package app

import (
	"context"
	"database/sql"
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

func (q *UserQueries) GetUserByID(ctx context.Context, userID string) (*User, error) {
	user, err := q.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (q *UserQueries) ListUsers(ctx context.Context) ([]*User, error) {
	users, err := q.usersRepo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}
