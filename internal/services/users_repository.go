package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	_ "modernc.org/sqlite" // SQLite driver
)

type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UsersRepository interface {
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
}

type sqliteUsersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) UsersRepository {
	return &sqliteUsersRepository{db: db}
}

func (r *sqliteUsersRepository) initSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *sqliteUsersRepository) CreateUser(ctx context.Context, user *User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.Must(uuid.NewV4()).String()
	}

	// Set timestamps
	now := time.Now()
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

func (r *sqliteUsersRepository) UpdateUser(ctx context.Context, user *User) error {
	// Update updated_at timestamp to current time
	user.UpdatedAt = time.Now()

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
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *sqliteUsersRepository) GetUserByID(ctx context.Context, userID string) (*User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *sqliteUsersRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at
		FROM users
		WHERE email = ?
	`
	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *sqliteUsersRepository) ListUsers(ctx context.Context) ([]*User, error) {
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

	var users []*User
	for rows.Next() {
		user := &User{}
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
