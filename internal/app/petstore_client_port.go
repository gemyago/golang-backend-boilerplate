/*
Package app contains application layer ports and services used by the Users & Pets example.

This package defines interfaces (ports) the application expects from infrastructure,
as well as concrete application services (commands/queries). Files in internal/app
should follow the "accept interface, return struct" pattern described in the project plan.
*/
package app

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
)

// PetstoreClient is the port (interface) for interacting with Petstore API.
// Defined in the application layer, it describes the capabilities needed by the app.
// The existing *petstore.Client struct satisfies this interface structurally.
// The application layer uses the concrete parameter types from the petstore package
// (AddPetParams, GetPetByIDParams) to match the infrastructure client signatures.
type PetstoreClient interface {
	AddPet(ctx context.Context, params petstore.AddPetParams) (*petstore.Pet, error)
	GetPetByID(ctx context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error)
}
