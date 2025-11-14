package app

import (
	"context"
	"errors"
	"log/slog"

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

func (q *PetsQueries) ListUserPets(_ context.Context, _ string) ([]*petstore.Pet, error) {
	return nil, errors.New("not implemented")
}
