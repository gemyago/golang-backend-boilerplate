package controllers

import (
	"log/slog"
	"testing"

	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ToolRegistrar defines the interface for registering tools with the server.
type ToolRegistrar interface {
	AddTools(tools ...server.ServerTool)
}

// MockTimeController for testing registry controller.
type MockTimeController struct {
	mock.Mock
}

func (m *MockTimeController) RegisterWithServer(server ToolRegistrar) error {
	args := m.Called(server)
	return args.Error(0)
}

// MockMathController for testing registry controller.
type MockMathController struct {
	mock.Mock
}

func (m *MockMathController) RegisterWithServer(server ToolRegistrar) error {
	args := m.Called(server)
	return args.Error(0)
}

func makeControllersRegistryDeps() ControllersRegistryDeps {
	return ControllersRegistryDeps{
		RootLogger:     slog.Default(),
		MathController: NewMathController(makeMathControllerDeps()),
		TimeController: NewTimeController(makeTimeControllerDeps()),
	}
}

func TestNewControllersRegistry(t *testing.T) {
	t.Run("should create registry with all controllers initialized", func(t *testing.T) {
		deps := makeControllersRegistryDeps()

		registry := NewControllersRegistry(deps)

		require.NotNil(t, registry)
		require.NotNil(t, registry.logger)
		require.NotNil(t, registry.mathController)
		require.NotNil(t, registry.timeController)
	})
}

func TestRegistry_RegisterAllControllers(t *testing.T) {
	t.Run("should register all controllers successfully", func(t *testing.T) {
		deps := makeControllersRegistryDeps()
		registry := NewControllersRegistry(deps)

		mockRegistrar := &MockToolRegistrar{}
		mockRegistrar.On("AddTools", mock.Anything).Return(nil).Times(2) // One for math, one for time

		err := registry.RegisterAllControllers(mockRegistrar)

		require.NoError(t, err)
		mockRegistrar.AssertExpectations(t)
	})
}

func TestRegistry_GetTimeController(t *testing.T) {
	t.Run("should return time controller instance", func(t *testing.T) {
		deps := makeControllersRegistryDeps()
		registry := NewControllersRegistry(deps)

		controller := registry.GetTimeController()

		require.NotNil(t, controller)
		assert.IsType(t, &TimeController{}, controller)
	})
}

func TestRegistry_GetMathController(t *testing.T) {
	t.Run("should return math controller instance", func(t *testing.T) {
		deps := makeControllersRegistryDeps()
		registry := NewControllersRegistry(deps)

		controller := registry.GetMathController()

		require.NotNil(t, controller)
		assert.IsType(t, &MathController{}, controller)
	})
}
