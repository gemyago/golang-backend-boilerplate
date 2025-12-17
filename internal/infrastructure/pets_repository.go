package infrastructure

import (
	"context"
	"database/sql"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"
	"go.uber.org/dig"
	_ "modernc.org/sqlite" // SQLite driver
)

type sqlitePetsRepository struct {
	db   *sql.DB
	time apptime.Provider
}

// Ensure sqlitePetsRepository implements app.PetsRepository.
var _ app.PetsRepository = (*sqlitePetsRepository)(nil)

type petsRepositoryDeps struct {
	dig.In

	DB   *Database
	Time apptime.Provider
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

func (r *sqlitePetsRepository) RemoveUserPet(ctx context.Context, userID string, petID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_pets
		WHERE user_id = ? AND pet_id = ?
	`, userID, petID)
	return err
}

func (r *sqlitePetsRepository) GetUserPetIDs(ctx context.Context, userID string) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pet_id FROM user_pets
		WHERE user_id = ?
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var petIDs []int64
	for rows.Next() {
		var petID int64
		if scanErr := rows.Scan(&petID); scanErr != nil {
			return nil, scanErr
		}
		petIDs = append(petIDs, petID)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return petIDs, nil
}

func (r *sqlitePetsRepository) HasUserPet(ctx context.Context, userID string, petID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_pets
			WHERE user_id = ? AND pet_id = ?
		)
	`, userID, petID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
