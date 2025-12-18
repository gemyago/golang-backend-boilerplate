package v1controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPets(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func(t *testing.T) PetsControllerDeps {
		mockCommands := NewMockPetsCommands(t)
		mockQueries := NewMockPetsQueries(t)
		deps := PetsControllerDeps{
			RootLogger:   telemetry.RootTestLogger(),
			PetsCommands: mockCommands,
			PetsQueries:  mockQueries,
		}
		return deps
	}
	newHandler := func(deps PetsControllerDeps) http.Handler {
		return server.NewTestRootHandler().
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
			deps := makeMockDeps(t)
			mockCmd := deps.PetsCommands.(*MockPetsCommands)
			petID := fake.Int64()
			mockCmd.EXPECT().AddPet(mock.Anything, app.AddPetRequest{
				UserID:    userID,
				Name:      payload.Name,
				Status:    string(payload.Status),
				PhotoUrls: payload.PhotoUrls,
			}).Return(&app.AddPetResponse{PetID: petID}, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code)
			var resp models.AddPetResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, petID, resp.PetID)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
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
			deps := makeMockDeps(t)
			mockCmd := deps.PetsCommands.(*MockPetsCommands)
			mockCmd.EXPECT().AddPet(mock.Anything, app.AddPetRequest{
				UserID:    userID,
				Name:      payload.Name,
				Status:    string(payload.Status),
				PhotoUrls: payload.PhotoUrls,
			}).Return(nil, errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("DELETE /users/{userId}/pets/{petId}", func(t *testing.T) {
		t.Run("happy path: returns 204", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID := fake.Int64()
			req := httptest.NewRequest(
				http.MethodDelete,
				"/users/"+userID+"/pets/"+strconv.FormatInt(petID, 10),
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.PetsCommands.(*MockPetsCommands)
			mockCmd.EXPECT().RemovePet(mock.Anything, userID, petID).Return(nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusNoContent, w.Code)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			userID := fake.UUID().V4()
			petID := fake.Int64()
			req := httptest.NewRequest(
				http.MethodDelete,
				"/users/"+userID+"/pets/"+strconv.FormatInt(petID, 10),
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.PetsCommands.(*MockPetsCommands)
			mockCmd.EXPECT().RemovePet(mock.Anything, userID, petID).Return(errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
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
			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.PetsQueries.(*MockPetsQueries)
			mockQueries.EXPECT().ListUserPets(mock.Anything, userID).Return(pets, nil)
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

		t.Run("internal error: returns 500", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.PetsQueries.(*MockPetsQueries)
			mockQueries.EXPECT().ListUserPets(mock.Anything, userID).Return(nil, errors.New(fake.Lorem().Sentence(1)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("empty: returns 200 with empty array", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/pets", nil)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.PetsQueries.(*MockPetsQueries)
			mockQueries.EXPECT().ListUserPets(mock.Anything, userID).Return([]*petstore.Pet{}, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUserPetsResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Empty(t, resp.Pets)
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
