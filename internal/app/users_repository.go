package app

import (
	"context"
	"time"
)

// User represents a user entity with all attributes needed for persistence and business logic.
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UsersRepository is the port (interface) for user persistence operations.
// Defined in application layer, implemented by infrastructure layer.
type UsersRepository interface {
	CreateUser(ctx context.Context, user User) error
	UpdateUser(ctx context.Context, user User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}
