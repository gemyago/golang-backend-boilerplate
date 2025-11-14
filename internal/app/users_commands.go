package app

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/dig"
)

// CreateUserRequest represents a request to create a new user.
type CreateUserRequest struct {
	Name  string
	Email string
}

type CreateUserResponse struct {
	UserID string
}

type UpdateUserRequest struct {
	UserID string
	Name   string
	Email  string
}

// Domain errors.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserEmailConflict = errors.New("user with this email already exists")
	ErrInvalidInput      = errors.New("invalid input")
)

// UserCommands is a concrete struct (not an interface).
// Controllers use this directly.
type UserCommands struct {
	usersRepo UsersRepository
	logger    *slog.Logger
}

type UserCommandsDeps struct {
	dig.In

	UsersRepo  UsersRepository
	RootLogger *slog.Logger
}

// NewUserCommands returns a concrete struct (not an interface).
// This follows "accept interface, return struct" principle.
func NewUserCommands(deps UserCommandsDeps) *UserCommands {
	return &UserCommands{
		usersRepo: deps.UsersRepo,
		logger:    deps.RootLogger.WithGroup("app.user-commands"),
	}
}

func (c *UserCommands) CreateUser(_ context.Context, _ CreateUserRequest) (*CreateUserResponse, error) {
	// Validate input
	// Check email uniqueness
	// Generate UUID
	// Create user in repository
	// Return user ID
	return nil, errors.New("not implemented")
}

func (c *UserCommands) UpdateUser(_ context.Context, _ UpdateUserRequest) error {
	// Validate input
	// Check user exists
	// Check email uniqueness (if changed)
	// Update user
	return errors.New("not implemented")
}

func (c *UserCommands) DeleteUser(_ context.Context, _ string) error {
	// Check user exists
	// Delete user (cascade will delete relationships)
	return errors.New("not implemented")
}
