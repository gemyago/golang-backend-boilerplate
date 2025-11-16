package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
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

func (c *UserCommands) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Normalize input
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	// Basic validation
	if req.Name == "" {
		return nil, NewErrInvalidInput("name", "cannot be empty")
	}
	if req.Email == "" {
		return nil, NewErrInvalidInput("email", "cannot be empty")
	}
	emailRe := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRe.MatchString(req.Email) {
		return nil, NewErrInvalidInput("email", "invalid format")
	}

	// Check email uniqueness
	existing, err := c.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// unexpected repository error
		return nil, err
	}
	if existing != nil {
		return nil, NewErrConflict("user email", "already exists")
	}

	// Generate UUID and create user
	id := uuid.Must(uuid.NewV4()).String()
	now := time.Now().UTC()
	user := User{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err = c.usersRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &CreateUserResponse{UserID: id}, nil
}

func (c *UserCommands) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	// Normalize input
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	// Basic validation
	if req.Name == "" {
		return NewErrInvalidInput("name", "cannot be empty")
	}
	if req.Email == "" {
		return NewErrInvalidInput("email", "cannot be empty")
	}
	emailRe := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRe.MatchString(req.Email) {
		return NewErrInvalidInput("email", "invalid format")
	}

	// Check user exists
	existing, err := c.usersRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewErrNotFound("user", req.UserID)
		}
		return err
	}

	// Check email uniqueness if email changed
	if existing.Email != req.Email {
		emailCheck, emailErr := c.usersRepo.GetUserByEmail(ctx, req.Email)
		if emailErr != nil && !errors.Is(emailErr, sql.ErrNoRows) {
			return emailErr
		}
		if emailCheck != nil {
			return NewErrConflict("user email", "already exists")
		}
	}

	// Create updated user entity
	updatedUser := User{
		ID:        req.UserID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: existing.CreatedAt, // Preserve original creation time
		UpdatedAt: time.Now().UTC(),   // Update timestamp
	}

	// Update user in repository
	return c.usersRepo.UpdateUser(ctx, updatedUser)
}

func (c *UserCommands) DeleteUser(ctx context.Context, userID string) error {
	// Verify user exists
	_, err := c.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewErrNotFound("user", userID)
		}
		return err
	}

	// Delete user; schema uses ON DELETE CASCADE for user_pets relationships
	err = c.usersRepo.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
