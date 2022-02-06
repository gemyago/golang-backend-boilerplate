package healthcheckv1

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T)
	}
	tests := []func() testCase{
		func() testCase {
			return testCase{
				name: "should respond with PONG",
				run: func(t *testing.T) {
					req := httptest.NewRequest("GET", "/ping", nil)
					recorder := httptest.NewRecorder()
					NewRoutes().ServeHTTP(recorder, req)
					assert.Equal(t, 200, recorder.Code)
					assert.Equal(t, "PONG", recorder.Body.String())
				},
			}
		},
	}
	for _, tt := range tests {
		tt := tt()
		t.Run(tt.name, tt.run)
	}
}
