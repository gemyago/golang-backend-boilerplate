package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/apptime"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"go.uber.org/dig"
	_ "modernc.org/sqlite" // SQLite driver
)

type SQLiteUsersRepository struct {
	db    *sql.DB
	time  apptime.Provider
	idGen ident.Generator
}

// Ensure sqliteUsersRepository implements app.UsersRepository.
var _ app.UsersRepository = (*SQLiteUsersRepository)(nil)

type UsersRepositoryDeps struct {
	dig.In

	DB    *Database
	Time  apptime.Provider
	IDGen ident.Generator
}

func NewUsersRepository(deps UsersRepositoryDeps) *SQLiteUsersRepository {
	return &SQLiteUsersRepository{
		db:    deps.DB.instance,
		time:  deps.Time,
		idGen: deps.IDGen,
	}
}

func (r *SQLiteUsersRepository) ensureRowsUpdated(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SQLiteUsersRepository) CreateUser(ctx context.Context, user app.User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = r.idGen.MustNewV7().String()
	}

	// Set timestamps
	now := r.time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Insert user
	query := `
		INSERT INTO users (id, name, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *SQLiteUsersRepository) UpdateUser(ctx context.Context, user app.User) error {
	// Update updated_at timestamp to current time
	user.UpdatedAt = r.time.Now()

	// Update user
	query := `
		UPDATE users
		SET name = ?, email = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return err
	}

	// Verify user exists (check affected rows)
	err = r.ensureRowsUpdated(result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app.NewErrNotFound("user", user.ID)
		}
		return err
	}
	return nil
}

func (r *SQLiteUsersRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `
		DELETE FROM users
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	err = r.ensureRowsUpdated(result)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app.NewErrNotFound("user", userID)
		}
		return err
	}
	return nil
}

func (r *SQLiteUsersRepository) GetUserByID(ctx context.Context, userID string) (*app.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &app.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.NewErrNotFound("user", userID)
		}
		return nil, err
	}
	return user, nil
}

func (r *SQLiteUsersRepository) GetUserByEmail(ctx context.Context, email string) (*app.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	user := &app.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, app.NewErrNotFound("user", email)
		}
		return nil, err
	}
	return user, nil
}
