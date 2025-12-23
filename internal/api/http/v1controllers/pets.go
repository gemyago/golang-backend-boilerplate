package v1controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

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

type PetsController struct {
	commands PetsCommands
	queries  PetsQueries
}

type PetsControllerDeps struct {
	dig.In

	PetsCommands
	PetsQueries

	RootLogger *slog.Logger
}

func newPetsController(deps PetsControllerDeps) *PetsController {
	return &PetsController{
		commands: deps.PetsCommands,
		queries:  deps.PetsQueries,
	}
}

// Ensure PetsController implements handlers.PetsController.
var _ handlers.PetsController = (*PetsController)(nil)

func (c *PetsController) AddUserPet(
	builder handlers.HandlerBuilder[*models.AddUserPetParams, *models.AddPetResponse],
) http.Handler {
	return builder.HandleWith(
		func(ctx context.Context, params *models.AddUserPetParams) (*models.AddPetResponse, error) {
			appReq := app.AddPetRequest{
				UserID:    params.UserID,
				Name:      params.Payload.Name,
				Status:    string(params.Payload.Status),
				PhotoUrls: params.Payload.PhotoUrls,
			}
			res, err := c.commands.AddPet(ctx, appReq)
			if err != nil {
				return nil, err
			}
			return &models.AddPetResponse{PetID: res.PetID}, nil
		},
	)
}

func (c *PetsController) RemoveUserPet(
	builder handlers.NoResponseHandlerBuilder[*models.RemoveUserPetParams],
) http.Handler {
	return builder.HandleWith(func(ctx context.Context, params *models.RemoveUserPetParams) error {
		return c.commands.RemovePet(ctx, params.UserID, params.PetID)
	})
}

func (c *PetsController) ListUserPets(
	builder handlers.HandlerBuilder[*models.ListUserPetsParams, *models.ListUserPetsResponse],
) http.Handler {
	return builder.HandleWith(
		func(ctx context.Context, params *models.ListUserPetsParams) (*models.ListUserPetsResponse, error) {
			pets, err := c.queries.ListUserPets(ctx, params.UserID)
			if err != nil {
				return nil, err
			}
			respPets := make([]*models.PetResponse, len(pets))
			for i, pet := range pets {
				respPets[i] = &models.PetResponse{
					ID:        pet.ID,
					Name:      pet.Name,
					Status:    string(pet.Status),
					PhotoUrls: pet.PhotoUrls,
				}
			}
			return &models.ListUserPetsResponse{Pets: respPets}, nil
		},
	)
}
