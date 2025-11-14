package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"go.uber.org/dig"
	_ "modernc.org/sqlite" // SQLite driver
)

type sqlitePetsRepository struct {
	db   *sql.DB
	time TimeProvider
}

// Ensure sqlitePetsRepository implements app.PetsRepository.
var _ app.PetsRepository = (*sqlitePetsRepository)(nil)

type petsRepositoryDeps struct {
	dig.In

	DB   *Database
	Time TimeProvider
}

func newPetsRepository(deps petsRepositoryDeps) *sqlitePetsRepository {
	return &sqlitePetsRepository{db: deps.DB.instance, time: deps.Time}
}

func (r *sqlitePetsRepository) AddUserPet(ctx context.Context, userPet app.UserPet) error {
	userPet.CreatedAt = r.time.Now()
	_, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO user_pets (user_id, pet_id, created_at)
		VALUES (?, ?, ?)
	`, userPet.UserID, userPet.PetID, userPet.CreatedAt)
	return err
}

func (r *sqlitePetsRepository) RemoveUserPet(_ context.Context, _ string, _ int64) error {
	return errors.New("not implemented")
}

func (r *sqlitePetsRepository) GetUserPetIDs(_ context.Context, _ string) ([]int64, error) {
	return nil, errors.New("not implemented")
}

func (r *sqlitePetsRepository) HasUserPet(_ context.Context, _ string, _ int64) (bool, error) {
	return false, errors.New("not implemented")
}
