package v1controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPetsController(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func(_ *testing.T, setupExpectations func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, mockPetstoreClient *MockPetstoreClient)) PetsControllerDeps {
		mockUsersRepo := &MockUsersRepository{}
		mockPetsRepo := &MockPetsRepository{}
		mockPetstoreClient := &MockPetstoreClient{}

		if setupExpectations != nil {
			setupExpectations(mockUsersRepo, mockPetsRepo, mockPetstoreClient)
		}

		commandsDeps := app.PetsCommandsDeps{
			PetsRepo:       mockPetsRepo,
			UsersRepo:      mockUsersRepo,
			PetstoreClient: mockPetstoreClient,
			RootLogger:     diag.RootTestLogger(),
		}
		commands := app.NewPetsCommands(commandsDeps)

		queriesDeps := app.PetsQueriesDeps{
			PetsRepo:       mockPetsRepo,
			UsersRepo:      mockUsersRepo,
			PetstoreClient: mockPetstoreClient,
			RootLogger:     diag.RootTestLogger(),
		}
		queries := app.NewPetsQueries(queriesDeps)

		return PetsControllerDeps{
			PetsCommands: commands,
			PetsQueries:  queries,
			RootLogger:   diag.RootTestLogger(),
		}
	}

	newHandler := func(deps PetsControllerDeps) http.Handler {
		return handlers.
			NewRootHandler(
				(*server.HTTPRouter)(http.NewServeMux()),
				handlers.WithLogger(deps.RootLogger),
				handlers.WithActionErrorHandler(func(w http.ResponseWriter, _ *http.Request, err error) {
					switch {
					case errors.Is(err, app.ErrInvalidInput):
						w.WriteHeader(http.StatusBadRequest)
					case errors.Is(err, app.ErrUserNotFound):
						w.WriteHeader(http.StatusNotFound)
					case errors.Is(err, app.ErrPetCreationFailed):
						w.WriteHeader(http.StatusBadGateway)
					case errors.Is(err, app.ErrUserPetNotFound):
						w.WriteHeader(http.StatusNotFound)
					default:
						w.WriteHeader(http.StatusInternalServerError)
					}
				}),
			).
			RegisterPetsRoutes(newPetsController(deps))
	}

	t.Run("POST /users/{userId}/pets", func(t *testing.T) {
		t.Run("happy path: returns 201 with petId", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := newRandomAddPetRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users/"+userID+"/pets",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			petID := fake.Int64()
			pet := &petstore.Pet{
				ID:        petID,
				Name:      payload.Name,
				Status:    petstore.PetStatus(payload.Status),
				PhotoUrls: payload.PhotoUrls,
			}
			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, mockPetstoreClient *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetstoreClient.On("AddPet", mock.Anything, mock.AnythingOfType("petstore.AddPetParams")).
						Return(pet, nil)
					mockPetsRepo.On("AddUserPet", mock.Anything, mock.AnythingOfType("app.UserPet")).Return(nil)
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code)
			var resp models.AddPetResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, petID, resp.PetID)
		})

		t.Run("validation error: returns 400 for invalid input", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := &models.AddPetRequest{
				Name:   "", // invalid
				Status: models.AddPetRequestStatus("available"),
			}
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users/"+userID+"/pets",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			deps := makeMockDeps(t, nil) // No expectations needed for validation

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("user not found: returns 404", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := newRandomAddPetRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users/"+userID+"/pets",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, _ *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return((*app.User)(nil), sql.ErrNoRows)
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("petstore failure: returns 502", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := newRandomAddPetRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users/"+userID+"/pets",
				bytes.NewBuffer(reqBody),
			)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, _ *MockPetsRepository, mockPetstoreClient *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetstoreClient.On("AddPet", mock.Anything, mock.AnythingOfType("petstore.AddPetParams")).
						Return((*petstore.Pet)(nil), errors.New("petstore error"))
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadGateway, w.Code)
		})
	})

	t.Run("DELETE /users/{userId}/pets/{petId}", func(t *testing.T) {
		t.Run("happy path: returns 204", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID := fake.Int64()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/users/%s/pets/%d", userID, petID),
				nil,
			)
			w := httptest.NewRecorder()

			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetsRepo.On("HasUserPet", mock.Anything, userID, petID).Return(true, nil)
					mockPetsRepo.On("RemoveUserPet", mock.Anything, userID, petID).Return(nil)
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusNoContent, w.Code)
		})

		t.Run("user not found: returns 404", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID := fake.Int64()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/users/%s/pets/%d", userID, petID),
				nil,
			)
			w := httptest.NewRecorder()

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, _ *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return((*app.User)(nil), sql.ErrNoRows)
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("relationship not found: returns 404", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID := fake.Int64()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/users/%s/pets/%d", userID, petID),
				nil,
			)
			w := httptest.NewRecorder()

			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetsRepo.On("HasUserPet", mock.Anything, userID, petID).Return(false, nil)
				},
			)

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	})

	t.Run("GET /users/{userId}/pets", func(t *testing.T) {
		t.Run("happy path: returns 200 with list of pets", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID1 := fake.Int64()
			petID2 := fake.Int64()
			pets := []*petstore.Pet{
				{
					ID:        petID1,
					Name:      fake.Lorem().Sentence(1),
					Status:    petstore.PetStatusAvailable,
					PhotoUrls: []string{fake.Internet().URL()},
				},
				{
					ID:        petID2,
					Name:      fake.Lorem().Sentence(1),
					Status:    petstore.PetStatusPending,
					PhotoUrls: []string{fake.Internet().URL(), fake.Internet().URL()},
				},
			}
			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, mockPetstoreClient *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetsRepo.On("GetUserPetIDs", mock.Anything, userID).Return([]int64{petID1, petID2}, nil)
					mockPetstoreClient.On("GetPetByID", mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID1, 10)}).
						Return(pets[0], nil)
					mockPetstoreClient.On("GetPetByID", mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID2, 10)}).
						Return(pets[1], nil)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()

			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUserPetsResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Len(t, resp.Pets, 2)
			assert.Equal(t, petID1, resp.Pets[0].ID)
			assert.Equal(t, pets[0].Name, resp.Pets[0].Name)
			assert.Equal(t, string(pets[0].Status), resp.Pets[0].Status)
			assert.Equal(t, pets[0].PhotoUrls, resp.Pets[0].PhotoUrls)

			assert.Equal(t, petID2, resp.Pets[1].ID)
			assert.Equal(t, pets[1].Name, resp.Pets[1].Name)
			assert.Equal(t, string(pets[1].Status), resp.Pets[1].Status)
			assert.Equal(t, pets[1].PhotoUrls, resp.Pets[1].PhotoUrls)
		})

		t.Run("user not found: returns 404", func(t *testing.T) {
			userID := fake.UUID().V4()

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, _ *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return((*app.User)(nil), sql.ErrNoRows)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()

			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("empty: returns 200 with empty array", func(t *testing.T) {
			userID := fake.UUID().V4()
			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, _ *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetsRepo.On("GetUserPetIDs", mock.Anything, userID).Return([]int64{}, nil)
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()

			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUserPetsResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Empty(t, resp.Pets)
		})

		t.Run("missing pets: gracefully handles when pet doesn't exist in petstore", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID1 := fake.Int64()
			petID2 := fake.Int64()
			pet := &petstore.Pet{
				ID:        petID1,
				Name:      fake.Lorem().Sentence(1),
				Status:    petstore.PetStatusAvailable,
				PhotoUrls: []string{fake.Internet().URL()},
			}
			user := &app.User{ID: userID}

			deps := makeMockDeps(
				t,
				func(mockUsersRepo *MockUsersRepository, mockPetsRepo *MockPetsRepository, mockPetstoreClient *MockPetstoreClient) {
					mockUsersRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
					mockPetsRepo.On("GetUserPetIDs", mock.Anything, userID).Return([]int64{petID1, petID2}, nil)
					mockPetstoreClient.On("GetPetByID", mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID1, 10)}).
						Return(pet, nil)
					mockPetstoreClient.On("GetPetByID", mock.Anything, petstore.GetPetByIDParams{PetID: strconv.FormatInt(petID2, 10)}).
						Return((*petstore.Pet)(nil), errors.New("pet not found"))
				},
			)

			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()

			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUserPetsResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Len(t, resp.Pets, 1)
			assert.Equal(t, petID1, resp.Pets[0].ID)
			assert.Equal(t, pet.Name, resp.Pets[0].Name)
			assert.Equal(t, string(pet.Status), resp.Pets[0].Status)
			assert.Equal(t, pet.PhotoUrls, resp.Pets[0].PhotoUrls)
		})
	})
}

func newRandomAddPetRequest(fake faker.Faker) *models.AddPetRequest {
	return &models.AddPetRequest{
		Name:      fake.Lorem().Sentence(1),
		Status:    models.AddPetRequestStatus("available"),
		PhotoUrls: []string{fake.Internet().URL()},
	}
}

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) CreateUser(ctx context.Context, user app.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepository) UpdateUser(ctx context.Context, user app.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUsersRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUsersRepository) GetUserByID(ctx context.Context, userID string) (*app.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*app.User), args.Error(1)
}

func (m *MockUsersRepository) GetUserByEmail(ctx context.Context, email string) (*app.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*app.User), args.Error(1)
}

func (m *MockUsersRepository) ListUsers(ctx context.Context) ([]*app.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*app.User), args.Error(1)
}

type MockPetsRepository struct {
	mock.Mock
}

func (m *MockPetsRepository) AddUserPet(ctx context.Context, userPet app.UserPet) error {
	args := m.Called(ctx, userPet)
	return args.Error(0)
}

func (m *MockPetsRepository) RemoveUserPet(ctx context.Context, userID string, petID int64) error {
	args := m.Called(ctx, userID, petID)
	return args.Error(0)
}

func (m *MockPetsRepository) GetUserPetIDs(ctx context.Context, userID string) ([]int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockPetsRepository) HasUserPet(ctx context.Context, userID string, petID int64) (bool, error) {
	args := m.Called(ctx, userID, petID)
	return args.Bool(0), args.Error(1)
}

type MockPetstoreClient struct {
	mock.Mock
}

func (m *MockPetstoreClient) AddPet(ctx context.Context, params petstore.AddPetParams) (*petstore.Pet, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*petstore.Pet), args.Error(1)
}

func (m *MockPetstoreClient) GetPetByID(ctx context.Context, params petstore.GetPetByIDParams) (*petstore.Pet, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*petstore.Pet), args.Error(1)
}
