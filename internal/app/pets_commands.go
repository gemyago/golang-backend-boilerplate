package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
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

func (c *PetsCommands) AddPet(ctx context.Context, req AddPetRequest) (*AddPetResponse, error) {
	if req.Name == "" {
		return nil, ErrInvalidInput
	}

	_, getUserErr := c.usersRepo.GetUserByID(ctx, req.UserID)
	if getUserErr != nil {
		if errors.Is(getUserErr, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", getUserErr)
	}

	petReq := &petstore.Pet{
		Name:      req.Name,
		Status:    petstore.PetStatus(req.Status),
		PhotoUrls: req.PhotoUrls,
	}

	pet, addPetErr := c.petstoreClient.AddPet(ctx, petstore.AddPetParams{Request: petReq})
	if addPetErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrPetCreationFailed, addPetErr)
	}

	userPet := UserPet{
		UserID:    req.UserID,
		PetID:     pet.ID,
		CreatedAt: time.Now(),
	}

	if addUserPetErr := c.petsRepo.AddUserPet(ctx, userPet); addUserPetErr != nil {
		return nil, fmt.Errorf("failed to add user pet relationship: %w", addUserPetErr)
	}

	return &AddPetResponse{PetID: pet.ID}, nil
}

func (c *PetsCommands) RemovePet(_ context.Context, _ string, _ int64) error {
	// TODO: Implement RemovePet command
	return errors.New("not implemented")
}
