package v1controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	fake := faker.New()

	makeMockDeps := func(_ *testing.T) UsersControllerDeps {
		deps := UsersControllerDeps{
			RootLogger: diag.RootTestLogger(),
		}
		return deps
	}
	newHandler := func(deps UsersControllerDeps) http.Handler {
		return handlers.
			NewRootHandler(
				(*server.HTTPRouter)(http.NewServeMux()),
				handlers.WithLogger(deps.RootLogger),
			).
			RegisterUsersRoutes(NewUsersController(deps))
	}

	t.Run("POST /users", func(t *testing.T) {
		t.Run("should be stub", func(t *testing.T) {
			payload := newRandomCreateUserRequest(fake)
			reqBody, _ := json.Marshal(payload)
			req := httptest.NewRequest(
				http.MethodPost,
				"/users",
				bytes.NewBuffer(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeMockDeps(t)
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
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
		t.Run("should be stub", func(t *testing.T) {
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
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}
