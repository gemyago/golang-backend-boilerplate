package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jaswdr/faker"
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

	t.Run("CreateUser", func(t *testing.T) {
		fake := faker.New()

		t.Run("should create user and be retrievable", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			user := NewRandomUser(fake)

			// When
			err = repo.CreateUser(ctx, user)

			// Then
			require.NoError(t, err)
			assert.NotEmpty(t, user.ID)

			// Verify user was created in database
			var count int
			query := "SELECT COUNT(*) FROM users WHERE id = ?"
			err = db.QueryRowContext(ctx, query, user.ID).Scan(&count)
			require.NoError(t, err)
			assert.Equal(t, 1, count)
		})

		t.Run("should return error for duplicate email", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			email := fake.Internet().Email()
			user1 := NewRandomUser(fake, WithUserEmail(email))
			user2 := NewRandomUser(fake, WithUserEmail(email))

			// Create first user
			err = repo.CreateUser(ctx, user1)
			require.NoError(t, err)

			// When - try to create second user with same email
			err = repo.CreateUser(ctx, user2)

			// Then
			require.Error(t, err)
			// Should be a unique constraint error
			assert.Contains(t, err.Error(), "UNIQUE constraint failed")
		})

		t.Run("should set timestamps correctly", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			beforeCreate := time.Now()
			user := NewRandomUser(fake, WithUserTimestamps(time.Time{}, time.Time{}))

			// When
			err = repo.CreateUser(ctx, user)

			// Then
			require.NoError(t, err)

			// Verify timestamps are set in database
			var createdAt, updatedAt time.Time
			query := "SELECT created_at, updated_at FROM users WHERE id = ?"
			err = db.QueryRowContext(ctx, query, user.ID).Scan(&createdAt, &updatedAt)
			require.NoError(t, err)
			assert.True(t, createdAt.After(beforeCreate) || createdAt.Equal(beforeCreate))
			assert.True(t, updatedAt.After(beforeCreate) || updatedAt.Equal(beforeCreate))
			assert.Equal(t, createdAt, updatedAt)
		})
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

	t.Run("GetUserByID", func(t *testing.T) {
		fake := faker.New()

		t.Run("should retrieve existing user correctly with all fields", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			user := NewRandomUser(fake)
			err = repo.CreateUser(ctx, user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByID(ctx, user.ID)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user.ID, retrievedUser.ID)
			assert.Equal(t, user.Name, retrievedUser.Name)
			assert.Equal(t, user.Email, retrievedUser.Email)
			assert.True(t, user.CreatedAt.Equal(retrievedUser.CreatedAt))
			assert.True(t, user.UpdatedAt.Equal(retrievedUser.UpdatedAt))
		})

		t.Run("should return error for non-existent user", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByID(ctx, "non-existent-id")

			// Then
			require.Error(t, err)
			assert.Equal(t, sql.ErrNoRows, err)
			assert.Nil(t, retrievedUser)
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		fake := faker.New()

		t.Run("should find user by email", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			user := NewRandomUser(fake)
			err = repo.CreateUser(ctx, user)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, user.Email)

			// Then
			require.NoError(t, err)
			require.NotNil(t, retrievedUser)
			assert.Equal(t, user.ID, retrievedUser.ID)
			assert.Equal(t, user.Name, retrievedUser.Name)
			assert.Equal(t, user.Email, retrievedUser.Email)
			assert.True(t, user.CreatedAt.Equal(retrievedUser.CreatedAt))
			assert.True(t, user.UpdatedAt.Equal(retrievedUser.UpdatedAt))
		})

		t.Run("should return error for non-existent email", func(t *testing.T) {
			// Given
			db, err := sql.Open("sqlite", ":memory:")
			require.NoError(t, err)
			defer db.Close()

			repo := NewUsersRepository(db)
			err = repo.(*sqliteUsersRepository).initSchema(ctx)
			require.NoError(t, err)

			// When
			retrievedUser, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")

			// Then
			require.Error(t, err)
			assert.Equal(t, sql.ErrNoRows, err)
			assert.Nil(t, retrievedUser)
		})
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
