package app

import (
	"context"
	"time"
)

// UserPet represents a user-pet relationship.
type UserPet struct {
	UserID    string
	PetID     int64
	CreatedAt time.Time
}

// PetsRepository is the port (interface) for pet relationship persistence.
// Defined in application layer, implemented by infrastructure layer.
type PetsRepository interface {
	AddUserPet(ctx context.Context, userPet UserPet) error
	RemoveUserPet(ctx context.Context, userID string, petID int64) error
	GetUserPetIDs(ctx context.Context, userID string) ([]int64, error)
	HasUserPet(ctx context.Context, userID string, petID int64) (bool, error)
}
