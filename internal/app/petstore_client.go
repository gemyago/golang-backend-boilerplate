package app

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
)

// PetstoreClient is the interface for interacting with Petstore API.
type PetstoreClient interface {
	AddPet(ctx context.Context, params petstore.AddPetParams) (*petstore.Pet, error)
	GetPetByID(ctx context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error)
}
