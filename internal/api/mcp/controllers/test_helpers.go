package controllers

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/mock"
)

// MockToolRegistrar is a mock implementation of the tool registrar for testing.
type MockToolRegistrar struct {
	mock.Mock
}

func (m *MockToolRegistrar) AddTools(tools ...server.ServerTool) {
	m.Called(tools)
}
