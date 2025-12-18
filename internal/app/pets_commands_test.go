package app

import (
	"context"
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
)

func TestPetsCommands(t *testing.T) {
	fake := faker.New()
	makeMockDeps := func(t *testing.T) PetsCommandsDeps {
		return PetsCommandsDeps{
			PetsRepo:       NewMockPetsRepository(t),
			UsersRepo:      NewMockUsersRepository(t),
			PetstoreClient: NewMockPetstoreClient(t),
			RootLogger:     telemetry.RootTestLogger(),
		}
	}

	t.Run("NewPetsCommands", func(t *testing.T) {
		t.Run("should create pets commands structure", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)

			// When
			commands := NewPetsCommands(deps)

			// Then
			require.NotNil(t, commands)
		})
	})

	t.Run("AddPet", func(t *testing.T) {
		t.Run("should create pet and add relationship", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := NewRandomAddPetRequest(fake)

			user := &User{ID: req.UserID}
			petID := int64(123)
			pet := &petstore.Pet{
				ID:        petID,
				Name:      req.Name,
				Status:    petstore.PetStatus(req.Status),
				PhotoUrls: req.PhotoUrls,
			}

			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(user, nil)
			mockPetstoreClient.EXPECT().AddPet(mock.Anything, mock.MatchedBy(func(params petstore.AddPetParams) bool {
				if params.Request == nil {
					return false
				}
				return params.Request.ID != 0 &&
					params.Request.Name == req.Name &&
					string(params.Request.Status) == req.Status &&
					len(params.Request.PhotoUrls) == len(req.PhotoUrls)
			})).Return(pet, nil)
			mockPetsRepo.EXPECT().AddUserPet(mock.Anything, mock.MatchedBy(func(up UserPet) bool {
				return up.UserID == req.UserID &&
					up.PetID == petID &&
					!up.CreatedAt.IsZero()
			})).Return(nil)

			// When
			resp, err := commands.AddPet(ctx, *req)

			// Then
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, petID, resp.PetID)
		})

		t.Run("should return ErrInvalidInput for empty name", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := NewRandomAddPetRequest(fake)
			req.Name = ""

			// When
			resp, err := commands.AddPet(ctx, *req)

			// Then
			require.Error(t, err)
			var errInvalidInput *InvalidInputError
			require.ErrorAs(t, err, &errInvalidInput)
			assert.Equal(t, "name", errInvalidInput.Field)
			require.Nil(t, resp)
		})

		t.Run("should return ErrUserNotFound when user doesn't exist", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := NewRandomAddPetRequest(fake)

			mockUsersRepo.EXPECT().
				GetUserByID(mock.Anything, req.UserID).
				Return((*User)(nil), NewErrNotFound("user", req.UserID))

			// When
			resp, err := commands.AddPet(ctx, *req)

			// Then
			require.Error(t, err)
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user", errNotFound.Resource)
			require.Nil(t, resp)
		})

		t.Run("should return ErrPetCreationFailed when petstore creation fails", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetstoreClient := deps.PetstoreClient.(*MockPetstoreClient)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := NewRandomAddPetRequest(fake)

			user := &User{ID: req.UserID}
			mockErr := errors.New("petstore boom")

			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return(user, nil)
			mockPetstoreClient.EXPECT().AddPet(mock.Anything, mock.Anything).Return((*petstore.Pet)(nil), mockErr)

			// When
			resp, err := commands.AddPet(ctx, *req)

			// Then
			require.Error(t, err)
			// Pet creation failures are now wrapped, check the error message
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to create pet in petstore")
			require.Nil(t, resp)
		})

		t.Run("should propagate unexpected user repo error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			req := NewRandomAddPetRequest(fake)

			mockErr := errors.New("unexpected db error")

			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, req.UserID).Return((*User)(nil), mockErr)

			// When
			resp, err := commands.AddPet(ctx, *req)

			// Then
			require.Error(t, err)
			require.ErrorIs(t, err, mockErr)
			require.Nil(t, resp)
		})
	})

	t.Run("RemovePet", func(t *testing.T) {
		t.Run("should remove user-pet relationship", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)
			mockPetsRepo.EXPECT().HasUserPet(mock.Anything, userID, petID).Return(true, nil)
			mockPetsRepo.EXPECT().RemoveUserPet(mock.Anything, userID, petID).Return(nil)

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.NoError(t, err)
		})

		t.Run("should return ErrUserNotFound when user doesn't exist", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			mockUsersRepo.EXPECT().
				GetUserByID(mock.Anything, userID).
				Return((*User)(nil), NewErrNotFound("user", userID))

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user", errNotFound.Resource)
		})

		t.Run("should return ErrUserPetNotFound when relationship doesn't exist", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			user := &User{ID: userID}
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)
			mockPetsRepo.EXPECT().HasUserPet(mock.Anything, userID, petID).Return(false, nil)

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			var errNotFound *NotFoundError
			require.ErrorAs(t, err, &errNotFound)
			assert.Equal(t, "user-pet relationship", errNotFound.Resource)
		})

		t.Run("should propagate unexpected user repo error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			mockErr := errors.New("unexpected db error")
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return((*User)(nil), mockErr)

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			require.ErrorIs(t, err, mockErr)
		})

		t.Run("should propagate unexpected pets repo error", func(t *testing.T) {
			// Given
			deps := makeMockDeps(t)
			mockUsersRepo := deps.UsersRepo.(*MockUsersRepository)
			mockPetsRepo := deps.PetsRepo.(*MockPetsRepository)
			commands := NewPetsCommands(deps)
			ctx := context.Background()
			userID := fake.UUID().V4()
			petID := fake.Int64Between(1, 1000)

			user := &User{ID: userID}
			mockErr := errors.New("unexpected db error")
			mockUsersRepo.EXPECT().GetUserByID(mock.Anything, userID).Return(user, nil)
			mockPetsRepo.EXPECT().HasUserPet(mock.Anything, userID, petID).Return(true, nil)
			mockPetsRepo.EXPECT().RemoveUserPet(mock.Anything, userID, petID).Return(mockErr)

			// When
			err := commands.RemovePet(ctx, userID, petID)

			// Then
			require.Error(t, err)
			require.ErrorIs(t, err, mockErr)
		})
	})
}
