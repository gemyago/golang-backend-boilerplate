package v1controllers

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
)

type UserCommands interface {
	CreateUser(ctx context.Context, req app.CreateUserRequest) (*app.CreateUserResponse, error)
	UpdateUser(ctx context.Context, req app.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID string) error
}

// Ensure interface is compatible.
var _ UserCommands = (*app.UserCommands)(nil)

type UserQueries interface {
	GetUserByID(ctx context.Context, userID string) (*app.User, error)
	ListUsers(ctx context.Context) ([]*app.User, error)
}

// Ensure interface is compatible.
var _ UserQueries = (*app.UserQueries)(nil)
