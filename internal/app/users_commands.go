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

func (c *UserCommands) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Normalize input
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	// Basic validation
	if req.Name == "" || req.Email == "" {
		return nil, ErrInvalidInput
	}
	emailRe := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRe.MatchString(req.Email) {
		return nil, ErrInvalidInput
	}

	// Check email uniqueness
	existing, err := c.usersRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// unexpected repository error
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserEmailConflict
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
