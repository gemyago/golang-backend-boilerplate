package v1controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func(t *testing.T) UsersControllerDeps {
		mockCommands := NewMockUserCommands(t)
		mockQueries := NewMockUserQueries(t)
		deps := UsersControllerDeps{
			RootLogger:   telemetry.RootTestLogger(),
			UserCommands: mockCommands,
			UserQueries:  mockQueries,
		}
		return deps
	}
	newHandler := func(deps UsersControllerDeps) http.Handler {
		return server.NewTestRootHandler().
			RegisterUsersRoutes(newUsersController(deps))
	}

	t.Run("POST /users", func(t *testing.T) {
		t.Run("happy path: returns 201 with userId", func(t *testing.T) {
			payload := newRandomCreateUserRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users",
				bytes.NewBuffer(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			userID := fake.UUID().V4()
			mockCmd.EXPECT().CreateUser(mock.Anything, app.CreateUserRequest{
				Name:  payload.Name,
				Email: payload.Email,
			}).Return(&app.CreateUserResponse{UserID: userID}, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code)
			var resp models.CreateUserResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, userID, resp.UserID)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			payload := newRandomCreateUserRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users",
				bytes.NewBuffer(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			mockCmd.EXPECT().CreateUser(mock.Anything, app.CreateUserRequest{
				Name:  payload.Name,
				Email: payload.Email,
			}).Return(nil, errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("DELETE /users/{userId}", func(t *testing.T) {
		t.Run("happy path: returns 204", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(
				http.MethodDelete,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			mockCmd.EXPECT().DeleteUser(mock.Anything, userID).Return(nil)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(
				http.MethodDelete,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			mockCmd.EXPECT().DeleteUser(mock.Anything, userID).Return(errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("GET /users/{userId}", func(t *testing.T) {
		t.Run("happy path: returns 200 with user data", func(t *testing.T) {
			userID := fake.UUID().V4()
			user := app.User{
				ID:    userID,
				Name:  fake.Person().Name(),
				Email: fake.Internet().Email(),
			}
			req := httptest.NewRequest(
				http.MethodGet,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.UserQueries.(*MockUserQueries)
			mockQueries.EXPECT().GetUserByID(mock.Anything, userID).Return(&user, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.UserResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, user.ID, resp.ID)
			assert.Equal(t, user.Name, resp.Name)
			assert.Equal(t, user.Email, resp.Email)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(
				http.MethodGet,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.UserQueries.(*MockUserQueries)
			mockQueries.EXPECT().GetUserByID(mock.Anything, userID).Return(nil, errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("GET /users", func(t *testing.T) {
		t.Run("happy path: returns 200 with list of users", func(t *testing.T) {
			users := []*app.User{
				{
					ID:    fake.UUID().V4(),
					Name:  fake.Person().Name(),
					Email: fake.Internet().Email(),
				},
				{
					ID:    fake.UUID().V4(),
					Name:  fake.Person().Name(),
					Email: fake.Internet().Email(),
				},
			}
			req := httptest.NewRequest(
				http.MethodGet,
				"/users",
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.UserQueries.(*MockUserQueries)
			mockQueries.EXPECT().ListUsers(mock.Anything).Return(users, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUsersResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.Len(t, resp.Users, 2)
			assert.Equal(t, users[0].ID, resp.Users[0].ID)
			assert.Equal(t, users[0].Name, resp.Users[0].Name)
			assert.Equal(t, users[0].Email, resp.Users[0].Email)
			assert.Equal(t, users[1].ID, resp.Users[1].ID)
			assert.Equal(t, users[1].Name, resp.Users[1].Name)
			assert.Equal(t, users[1].Email, resp.Users[1].Email)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/users",
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.UserQueries.(*MockUserQueries)
			mockQueries.EXPECT().ListUsers(mock.Anything).Return(nil, errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("empty: returns 200 with empty array", func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/users",
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockQueries := deps.UserQueries.(*MockUserQueries)
			mockQueries.EXPECT().ListUsers(mock.Anything).Return([]*app.User{}, nil)
			newHandler(deps).ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			var resp models.ListUsersResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Empty(t, resp.Users)
		})
	})

	t.Run("PUT /users/{userId}", func(t *testing.T) {
		t.Run("happy path: returns 204", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := newRandomUpdateUserRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPut,
				"/users/"+userID,
				bytes.NewBuffer(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			mockCmd.EXPECT().UpdateUser(mock.Anything, app.UpdateUserRequest{
				UserID: userID,
				Name:   payload.Name,
				Email:  payload.Email,
			}).Return(nil)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
		})

		t.Run("internal error: returns 500", func(t *testing.T) {
			userID := fake.UUID().V4()
			payload := newRandomUpdateUserRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPut,
				"/users/"+userID,
				bytes.NewBuffer(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			mockCmd := deps.UserCommands.(*MockUserCommands)
			mockCmd.EXPECT().UpdateUser(mock.Anything, app.UpdateUserRequest{
				UserID: userID,
				Name:   payload.Name,
				Email:  payload.Email,
			}).Return(errors.New(fake.Lorem().Sentence(3)))
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
