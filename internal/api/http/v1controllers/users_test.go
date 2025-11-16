package v1controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func(t *testing.T) UsersControllerDeps {
		mockCommands := NewMockUserCommands(t)
		mockQueries := NewMockUserQueries(t)
		deps := UsersControllerDeps{
			RootLogger:   diag.RootTestLogger(),
			UserCommands: mockCommands,
			UserQueries:  mockQueries,
		}
		return deps
	}
	newHandler := func(deps UsersControllerDeps) http.Handler {
		return handlers.
			NewRootHandler(
				(*server.HTTPRouter)(http.NewServeMux()),
				handlers.WithLogger(deps.RootLogger),
				handlers.WithActionErrorHandler(func(w http.ResponseWriter, _ *http.Request, err error) {
					switch {
					case errors.Is(err, app.ErrInvalidInput):
						w.WriteHeader(http.StatusBadRequest)
					case errors.Is(err, app.ErrUserEmailConflict):
						w.WriteHeader(http.StatusConflict)
					case errors.Is(err, app.ErrUserNotFound):
						w.WriteHeader(http.StatusNotFound)
					default:
						w.WriteHeader(http.StatusInternalServerError)
					}
					// Log the error if needed, but for test, just set status
				}),
			).
			RegisterUsersRoutes(NewUsersController(deps))
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

		t.Run("validation error: returns 400 for invalid input", func(t *testing.T) {
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
			}).Return(nil, app.ErrInvalidInput)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("conflict: returns 409 for duplicate email", func(t *testing.T) {
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
			}).Return(nil, app.ErrUserEmailConflict)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code)
		})
	})

	t.Run("DELETE /users/{userId}", func(t *testing.T) {
		t.Run("should be stub", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(
				http.MethodDelete,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("GET /users/{userId}", func(t *testing.T) {
		t.Run("should be stub", func(t *testing.T) {
			userID := fake.UUID().V4()
			req := httptest.NewRequest(
				http.MethodGet,
				"/users/"+userID,
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("GET /users", func(t *testing.T) {
		t.Run("should be stub", func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				"/users",
				nil,
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
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

		t.Run("validation error: returns 400", func(t *testing.T) {
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
			}).Return(app.ErrInvalidInput)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("not found: returns 404", func(t *testing.T) {
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
			}).Return(app.ErrUserNotFound)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("conflict: returns 409", func(t *testing.T) {
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
			}).Return(app.ErrUserEmailConflict)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code)
		})
	})
}
