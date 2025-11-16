package v1controllers

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
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

type PetsCommands interface {
	AddPet(ctx context.Context, req app.AddPetRequest) (*app.AddPetResponse, error)
	RemovePet(ctx context.Context, userID string, petID int64) error
}

// Ensure interface is compatible.
var _ PetsCommands = (*app.PetsCommands)(nil)

type PetsQueries interface {
	ListUserPets(ctx context.Context, userID string) ([]*petstore.Pet, error)
}

// Ensure interface is compatible.
var _ PetsQueries = (*app.PetsQueries)(nil)
