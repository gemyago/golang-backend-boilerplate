package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/dig"
	_ "modernc.org/sqlite" // SQLite driver
)

type sqliteUsersRepository struct {
	db   *sql.DB
	time TimeProvider
}

// Ensure sqliteUsersRepository implements app.UsersRepository.
var _ app.UsersRepository = (*sqliteUsersRepository)(nil)

type usersRepositoryDeps struct {
	dig.In

	DB   *Database
	Time TimeProvider
}

func newUsersRepository(deps usersRepositoryDeps) *sqliteUsersRepository {
	return &sqliteUsersRepository{db: deps.DB.instance, time: deps.Time}
}

func (r *sqliteUsersRepository) ensureRowsUpdated(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *sqliteUsersRepository) CreateUser(ctx context.Context, user app.User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.Must(uuid.NewV4()).String()
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

func (r *sqliteUsersRepository) UpdateUser(ctx context.Context, user app.User) error {
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

func (r *sqliteUsersRepository) DeleteUser(ctx context.Context, userID string) error {
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

func (r *sqliteUsersRepository) GetUserByID(ctx context.Context, userID string) (*app.User, error) {
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

func (r *sqliteUsersRepository) GetUserByEmail(ctx context.Context, email string) (*app.User, error) {
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

func (r *sqliteUsersRepository) ListUsers(ctx context.Context) ([]*app.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*app.User
	for rows.Next() {
		user := &app.User{}
		scanErr := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return users, nil
}
