package ident

import (
	"fmt"

	uuid "github.com/gofrs/uuid/v5"
)

// UUID represents a universally unique identifier.
// Should be used in the application code instead of directly using uuid.UUID.
// Use it to minimize dependency on the underlying UUID library.
type UUID = uuid.UUID

// Not creating a new type to avoid unnecessary conversions. The gofrs defines
// everything related to marshalling, sql scanning, etc. on uuid.UUID type.

// Generator defines the contract for generating unique identifiers.
type Generator interface {
	// MustNewV7 generates a new UUIDv7 identifier or panics on failure.
	// Consider UUIDv7 function if error handling is required.
	MustNewV7() UUID

	// NewV7 generates a new NewV7 identifier and allows error handling.
	NewV7() (UUID, error)
}

type DefaultGenerator struct {
	gen *uuid.Gen
}

var _ Generator = (*DefaultGenerator)(nil)

// NewDefaultGenerator creates a new instance of Generator.
func NewDefaultGenerator() *DefaultGenerator {
	return &DefaultGenerator{
		gen: uuid.NewGen(),
	}
}

func (g *DefaultGenerator) MustNewV7() UUID {
	id, err := g.gen.NewV7()
	if err != nil { // coverage-ignore // no idea how to simulate this gracefully
		panic(fmt.Errorf("failed to generate UUIDv7: %w", err))
	}
	return id
}

func (g *DefaultGenerator) NewV7() (UUID, error) {
	return g.gen.NewV7()
}
