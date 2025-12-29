package app

import (
	"context"
	"fmt"
	"log/slog"

	"go.uber.org/dig"
)

// UserQueries is a concrete struct (not an interface).
// Controllers and other consumers use this directly.
// Follows "accept interface, return struct" principle.
type UserQueries struct {
	usersRepo       UsersRepository
	logger          *slog.Logger
	databaseQuerier Queryer
}

type UserQueriesDeps struct {
	dig.In

	Queryer

	UsersRepo  UsersRepository
	RootLogger *slog.Logger
}

// NewUserQueries returns a concrete struct (not an interface).
// This follows "accept interface, return struct" principle.
func NewUserQueries(deps UserQueriesDeps) *UserQueries {
	return &UserQueries{
		usersRepo:       deps.UsersRepo,
		logger:          deps.RootLogger.WithGroup("app.user-queries"),
		databaseQuerier: deps.Queryer,
	}
}

func (q *UserQueries) GetUserByID(ctx context.Context, userID string) (*User, error) {
	user, err := q.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (q *UserQueries) ListUsers(ctx context.Context) ([]*User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at ASC
	`
	rows, err := q.databaseQuerier.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		scanErr := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return users, nil
}
