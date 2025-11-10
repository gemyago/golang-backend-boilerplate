package v1controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	fake := faker.New()
	type mockDeps struct {
		EchoController handlers.EchoController
	}
	makeDeps := func() mockDeps {
		// In real world example a mock of EchoService would be used
		echoService := app.NewEchoService(app.EchoServiceDeps{
			RootLogger: diag.RootTestLogger(),
		})
		deps := mockDeps{
			EchoController: EchoController{EchoService: echoService},
		}
		return deps
	}
	t.Run("POST /echo", func(t *testing.T) {
		t.Run("should respond with OK", func(t *testing.T) {
			wantMessage := fake.Lorem().Sentence(10)
			reqBody := `{"message": "` + wantMessage + `"}`
			req := httptest.NewRequest(
				http.MethodPost,
				"/echo",
				bytes.NewBufferString(reqBody),
			)
			w := httptest.NewRecorder()
			deps := makeDeps()
			rootHandler := handlers.
				NewRootHandler((*server.HTTPRouter)(http.NewServeMux())).
				RegisterEchoRoutes(deps.EchoController)
			rootHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, `{"message":"`+wantMessage+`"}`, w.Body.String())
		})
	})
}
