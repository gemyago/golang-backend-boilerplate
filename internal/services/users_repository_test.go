package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("should initialize schema", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)

		// When
		err = repo.(*sqliteUsersRepository).initSchema(ctx)

		// Then
		require.NoError(t, err)

		// Verify table exists
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'"
		err = db.QueryRowContext(ctx, query).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("CreateUser should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)
		user := &User{}

		// When
		err = repo.CreateUser(ctx, user)

		// Then
		assert.EqualError(t, err, "not implemented")
	})

	t.Run("UpdateUser should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)
		user := &User{}

		// When
		err = repo.UpdateUser(ctx, user)

		// Then
		assert.EqualError(t, err, "not implemented")
	})

	t.Run("DeleteUser should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)

		// When
		err = repo.DeleteUser(ctx, "user-id")

		// Then
		assert.EqualError(t, err, "not implemented")
	})

	t.Run("GetUserByID should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)

		// When
		_, err = repo.GetUserByID(ctx, "user-id")

		// Then
		assert.EqualError(t, err, "not implemented")
	})

	t.Run("GetUserByEmail should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)

		// When
		_, err = repo.GetUserByEmail(ctx, "user@example.com")

		// Then
		assert.EqualError(t, err, "not implemented")
	})

	t.Run("ListUsers should return not implemented", func(t *testing.T) {
		// Given
		db, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		defer db.Close()

		repo := NewUsersRepository(db)

		// When
		_, err = repo.ListUsers(ctx)

		// Then
		assert.EqualError(t, err, "not implemented")
	})
}
