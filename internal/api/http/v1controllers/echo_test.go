package v1controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	fake := faker.New()
	type mockDeps struct {
		echoService *app.EchoService
	}
	makeMockDeps := func() mockDeps {
		rootLogger := telemetry.RootTestLogger()

		// In real world example a mock of EchoService would be used
		echoService := app.NewEchoService(app.EchoServiceDeps{
			RootLogger: rootLogger,
		})
		return mockDeps{
			echoService: echoService,
		}
	}
	newHandler := func(deps mockDeps) http.Handler {
		return server.NewTestRootHandler().
			RegisterEchoRoutes(newEchoController(deps.echoService))
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
			deps := makeMockDeps()
			newHandler(deps).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.JSONEq(t, `{"message":"`+wantMessage+`"}`, w.Body.String())
		})
	})
}
