package middleware

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockRoundTripper is a mock implementation of [http.RoundTripper] for testing.
type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	res, _ := args.Get(0).(*http.Response)
	return res, args.Error(1)
}
