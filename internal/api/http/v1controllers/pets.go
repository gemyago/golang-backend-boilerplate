package v1controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"go.uber.org/dig"
)

type PetsController struct {
	commands *app.PetsCommands
	queries  *app.PetsQueries
}

type PetsControllerDeps struct {
	dig.In

	PetsCommands *app.PetsCommands
	PetsQueries  *app.PetsQueries
	RootLogger   *slog.Logger
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
	return builder.HandleWith(func(_ context.Context, _ *models.RemoveUserPetParams) error {
		// TODO: implement
		return nil
	})
}

func (c *PetsController) ListUserPets(
	builder handlers.HandlerBuilder[*models.ListUserPetsParams, *models.ListUserPetsResponse],
) http.Handler {
	return builder.HandleWith(
		func(_ context.Context, _ *models.ListUserPetsParams) (*models.ListUserPetsResponse, error) {
			// TODO: implement
			return nil, nil
		},
	)
}
