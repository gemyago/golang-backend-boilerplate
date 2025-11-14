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

func (r *sqlitePetsRepository) AddUserPet(_ context.Context, _ app.UserPet) error {
	return errors.New("not implemented")
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
