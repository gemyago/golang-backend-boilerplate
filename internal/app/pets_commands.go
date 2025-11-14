package app

import (
	"context"
	"errors"
	"log/slog"

	"go.uber.org/dig"
)

// AddPetRequest represents a request to add a pet to a user.
type AddPetRequest struct {
	UserID    string
	Name      string
	Status    string
	PhotoUrls []string
}

type AddPetResponse struct {
	PetID int64
}

// Domain errors.
var (
	ErrPetCreationFailed = errors.New("failed to create pet in petstore")
	ErrUserPetNotFound   = errors.New("user-pet relationship not found")
)

// PetsCommands is a concrete struct (not an interface).
// Controllers use this directly.
type PetsCommands struct {
	petsRepo       PetsRepository
	usersRepo      UsersRepository
	petstoreClient PetstoreClient
	logger         *slog.Logger
}

type PetsCommandsDeps struct {
	dig.In

	PetsRepo       PetsRepository
	UsersRepo      UsersRepository
	PetstoreClient PetstoreClient
	RootLogger     *slog.Logger
}

// NewPetsCommands returns a concrete struct (not an interface).
// This follows "accept interface, return struct" principle.
func NewPetsCommands(deps PetsCommandsDeps) *PetsCommands {
	return &PetsCommands{
		petsRepo:       deps.PetsRepo,
		usersRepo:      deps.UsersRepo,
		petstoreClient: deps.PetstoreClient,
		logger:         deps.RootLogger.WithGroup("app.pets-commands"),
	}
}

func (c *PetsCommands) AddPet(_ context.Context, _ AddPetRequest) (*AddPetResponse, error) {
	// TODO: Implement AddPet command
	return nil, errors.New("not implemented")
}

func (c *PetsCommands) RemovePet(_ context.Context, _ string, _ int64) error {
	// TODO: Implement RemovePet command
	return errors.New("not implemented")
}
