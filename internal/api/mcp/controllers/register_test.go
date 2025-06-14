package controllers

import (
	"log/slog"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func makeRegistryDeps() RegistryDeps {
	// Create real services for dependency injection
	timeServiceDeps := app.TimeServiceDeps{
		RootLogger: slog.Default(),
	}
	timeService := app.NewTimeService(timeServiceDeps)

	mathServiceDeps := app.MathServiceDeps{
		RootLogger: slog.Default(),
	}
	mathService := app.NewMathService(mathServiceDeps)

	return RegistryDeps{
		RootLogger:  slog.Default(),
		TimeService: timeService,
		MathService: mathService,
	}
}

func TestNewControllersRegistry(t *testing.T) {
	t.Run("should create registry with all controllers initialized", func(t *testing.T) {
		deps := makeRegistryDeps()

		registry := NewControllersRegistry(deps)

		require.NotNil(t, registry)
		require.NotNil(t, registry.logger)
		require.NotNil(t, registry.timeController)
		require.NotNil(t, registry.mathController)
	})
}

func TestRegistry_RegisterAllControllers(t *testing.T) {
	t.Run("should register all controllers successfully", func(t *testing.T) {
		deps := makeRegistryDeps()
		registry := NewControllersRegistry(deps)
		ctx := t.Context()

		mockRegistrar := &MockToolRegistrar{}
		// For time controller
		mockRegistrar.On("RegisterTool", mock.AnythingOfType("mcp.Tool"), mock.AnythingOfType("server.ToolHandlerFunc")).
			Return(nil).Once()
		// For math controller
		mockRegistrar.On("AddTools", mock.Anything).Return().Once()

		err := registry.RegisterAllControllers(ctx, mockRegistrar)

		require.NoError(t, err)
		mockRegistrar.AssertExpectations(t)
	})
}

func TestRegistry_GetTimeController(t *testing.T) {
	t.Run("should return time controller instance", func(t *testing.T) {
		deps := makeRegistryDeps()
		registry := NewControllersRegistry(deps)

		timeController := registry.GetTimeController()

		require.NotNil(t, timeController)
		assert.Equal(t, registry.timeController, timeController)
	})
}

func TestRegistry_GetMathController(t *testing.T) {
	t.Run("should return math controller instance", func(t *testing.T) {
		deps := makeRegistryDeps()
		registry := NewControllersRegistry(deps)

		mathController := registry.GetMathController()

		require.NotNil(t, mathController)
		assert.Equal(t, registry.mathController, mathController)
	})
}
