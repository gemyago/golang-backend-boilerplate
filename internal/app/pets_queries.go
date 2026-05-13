package app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

// PetsQueries is a concrete struct (not an interface).
// Controllers and other consumers use this directly.
// Follows "accept interface, return struct" principle.
type PetsQueries struct {
	petsRepo       PetsRepository
	usersRepo      UsersRepository
	petstoreClient PetstoreClient
	logger         *slog.Logger
	tracer         trace.Tracer
	meeter         metric.Meter
}

type PetsQueriesDeps struct {
	dig.In

	TracerProvider trace.TracerProvider
	MeeterProvider metric.MeterProvider

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
		tracer:         deps.TracerProvider.Tracer("PetsQueries"),
		meeter:         deps.MeeterProvider.Meter("PetsQueries"),
	}
}

func (q *PetsQueries) ListUserPets(ctx context.Context, userID string) ([]*petstore.Pet, error) {
	ctx, span := q.tracer.Start(ctx, "ListUserPets")
	defer span.End()

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

	pets := make([]*petstore.Pet, 0, len(petIDs))

	fetchedPetsResult := make(chan *petstore.Pet, len(petIDs))

	// We don't want to flood the petstore. In real life this may be configurable.
	const maxConcurrentFetches = 3
	fetchGrp, grpCtx := errgroup.WithContext(ctx)
	fetchGrp.SetLimit(maxConcurrentFetches)

	for _, petID := range petIDs {
		fetchGrp.Go(func() error {
			pet, petErr := q.petstoreClient.GetPetByID(grpCtx, petstore.GetPetByIDParams{
				PetID: strconv.FormatInt(petID, 10),
			})
			if petErr != nil {
				// Log warning and skip missing pet
				q.logger.WarnContext(grpCtx,
					"failed to fetch pet details from petstore",
					slog.Int64("pet_id", petID),
					slog.String("error", petErr.Error()),
				)
			} else {
				fetchedPetsResult <- pet
			}
			return nil
		})
	}

	go func() {
		_ = fetchGrp.Wait()
		close(fetchedPetsResult)
	}()

	for pet := range fetchedPetsResult {
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
