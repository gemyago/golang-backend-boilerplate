package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

// PetsQueries is a concrete struct (not an interface).
// Controllers and other consumers use this directly.
// Follows "accept interface, return struct" principle.
type PetsQueries struct {
	petsRepo       PetsRepository
	usersRepo      UsersRepository
	petstoreClient PetstoreClient
	logger         *slog.Logger
}

type PetsQueriesDeps struct {
	dig.In

	PetsRepo       PetsRepository
	UsersRepo      UsersRepository
	PetstoreClient PetstoreClient
	RootLogger     *slog.Logger
}

// NewPetsQueries returns a concrete struct (not an interface).
// This follows "accept interface, return struct" principle.
func NewPetsQueries(deps PetsQueriesDeps) *PetsQueries {
	return &PetsQueries{
		petsRepo:       deps.PetsRepo,
		usersRepo:      deps.UsersRepo,
		petstoreClient: deps.PetstoreClient,
		logger:         deps.RootLogger.WithGroup("app.pets-queries"),
	}
}

func (q *PetsQueries) ListUserPets(ctx context.Context, userID string) ([]*petstore.Pet, error) {
	// Verify user exists
	_, err := q.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get pet IDs for the user
	petIDs, err := q.petsRepo.GetUserPetIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pet IDs: %w", err)
	}

	q.logger.InfoContext(ctx,
		"Resolving pet details from petstore",
		slog.String("user_id", userID),
		slog.Int("pet_count", len(petIDs)),
	)

	// Fetch pet details from Petstore for each ID
	var pets []*petstore.Pet
	for _, petID := range petIDs {
		pet, petErr := q.petstoreClient.GetPetByID(ctx, petstore.GetPetByIDParams{
			PetID: strconv.FormatInt(petID, 10),
		})
		if petErr != nil {
			// Log warning and skip missing pet
			q.logger.WarnContext(ctx,
				"failed to fetch pet details from petstore",
				slog.Int64("pet_id", petID),
				slog.String("error", petErr.Error()),
			)
			continue
		}
		pets = append(pets, pet)
	}

	q.logger.DebugContext(
		ctx,
		"Resolved details for user pets from petstore",
		slog.String("user_id", userID),
		slog.Int("resolved_pet_count", len(pets)),
	)

	return pets, nil
}
