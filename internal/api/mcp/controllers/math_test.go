package controllers

import (
	"log/slog"
	"testing"

	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockToolRegistrar is a mock implementation of the tool registrar for testing.
type MockToolRegistrar struct {
	mock.Mock
}

func (m *MockToolRegistrar) RegisterTool(tool mcp.Tool, handler ToolHandler) error {
	args := m.Called(tool, handler)
	return args.Error(0)
}

func makeMathControllerDeps() MathControllerDeps {
	mathServiceDeps := app.MathServiceDeps{
		RootLogger: slog.Default(),
	}
	mathService := app.NewMathService(mathServiceDeps)

	return MathControllerDeps{
		MathService: mathService,
		RootLogger:  slog.Default(),
	}
}

func TestNewMathController(t *testing.T) {
	t.Run("should create math controller with dependencies", func(t *testing.T) {
		deps := makeMathControllerDeps()

		controller := NewMathController(deps)

		require.NotNil(t, controller)
		require.NotNil(t, controller.mathService)
		require.NotNil(t, controller.logger)
	})
}

func TestMathController_ToolDefinitions(t *testing.T) {
	t.Run("should return calculate tool with correct schema", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		tool := controller.GetCalculateTool()

		assert.Equal(t, "calculate", tool.Name)
		assert.Equal(t, "Perform mathematical calculations (add, subtract, multiply, divide)", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})

	t.Run("should return add tool with correct schema", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		tool := controller.GetAddTool()

		assert.Equal(t, "add", tool.Name)
		assert.Equal(t, "Add two numbers together", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})

	t.Run("should return subtract tool with correct schema", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		tool := controller.GetSubtractTool()

		assert.Equal(t, "subtract", tool.Name)
		assert.Equal(t, "Subtract second number from first number", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})

	t.Run("should return multiply tool with correct schema", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		tool := controller.GetMultiplyTool()

		assert.Equal(t, "multiply", tool.Name)
		assert.Equal(t, "Multiply two numbers together", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})

	t.Run("should return divide tool with correct schema", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		tool := controller.GetDivideTool()

		assert.Equal(t, "divide", tool.Name)
		assert.Equal(t, "Divide first number by second number", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})
}

func TestMathController_HandleCalculate(t *testing.T) {
	t.Run("should handle calculate add request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "calculate",
				Arguments: map[string]interface{}{
					"operation": "add",
					"a":         5.0,
					"b":         3.0,
				},
			},
		}

		result, err := controller.HandleCalculate(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "Result: 8")
			assert.Contains(t, content.Text, "operation: add")
		}
	})

	t.Run("should handle calculate multiply request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "calculate",
				Arguments: map[string]interface{}{
					"operation": "multiply",
					"a":         6.0,
					"b":         7.0,
				},
			},
		}

		result, err := controller.HandleCalculate(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "Result: 42")
			assert.Contains(t, content.Text, "operation: multiply")
		}
	})

	t.Run("should handle invalid operation parameter", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "calculate",
				Arguments: map[string]interface{}{
					"operation": 123, // Invalid type
					"a":         5.0,
					"b":         3.0,
				},
			},
		}

		result, err := controller.HandleCalculate(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "Error content should be text content")
			assert.Contains(t, content.Text, "Invalid parameters")
		}
	})

	t.Run("should handle missing parameters", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "calculate",
				Arguments: map[string]interface{}{
					"operation": "add",
					"a":         5.0,
					// Missing "b" parameter
				},
			},
		}

		result, err := controller.HandleCalculate(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "Error content should be text content")
			assert.Contains(t, content.Text, "Invalid parameters")
		}
	})

	t.Run("should handle division by zero error", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "calculate",
				Arguments: map[string]interface{}{
					"operation": "divide",
					"a":         5.0,
					"b":         0.0,
				},
			},
		}

		result, err := controller.HandleCalculate(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "Error content should be text content")
			assert.Contains(t, content.Text, "Calculation failed")
		}
	})
}

func TestMathController_HandleAdd(t *testing.T) {
	t.Run("should handle add request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add",
				Arguments: map[string]interface{}{
					"a": 7.0,
					"b": 3.0,
				},
			},
		}

		result, err := controller.HandleAdd(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "7 + 3 = 10")
		}
	})

	t.Run("should handle invalid parameters", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "add",
				Arguments: map[string]interface{}{
					"a": "invalid", // Invalid type
					"b": 3.0,
				},
			},
		}

		result, err := controller.HandleAdd(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "Error content should be text content")
			assert.Contains(t, content.Text, "Invalid parameters")
		}
	})
}

func TestMathController_HandleSubtract(t *testing.T) {
	t.Run("should handle subtract request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "subtract",
				Arguments: map[string]interface{}{
					"a": 10.0,
					"b": 4.0,
				},
			},
		}

		result, err := controller.HandleSubtract(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "10 - 4 = 6")
		}
	})
}

func TestMathController_HandleMultiply(t *testing.T) {
	t.Run("should handle multiply request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "multiply",
				Arguments: map[string]interface{}{
					"a": 6.0,
					"b": 7.0,
				},
			},
		}

		result, err := controller.HandleMultiply(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "6 × 7 = 42")
		}
	})
}

func TestMathController_HandleDivide(t *testing.T) {
	t.Run("should handle divide request successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "divide",
				Arguments: map[string]interface{}{
					"a": 20.0,
					"b": 4.0,
				},
			},
		}

		result, err := controller.HandleDivide(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "First content should be text content")
			assert.Contains(t, content.Text, "20 ÷ 4 = 5")
		}
	})

	t.Run("should handle division by zero error", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)
		ctx := t.Context()

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "divide",
				Arguments: map[string]interface{}{
					"a": 10.0,
					"b": 0.0,
				},
			},
		}

		result, err := controller.HandleDivide(ctx, request)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.NotEmpty(t, result.Content)

		if len(result.Content) > 0 {
			content, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "Error content should be text content")
			assert.Contains(t, content.Text, "Division failed")
		}
	})
}

func TestMathController_ParameterExtraction(t *testing.T) {
	t.Run("should extract number parameters correctly", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		// Test with float64
		args := map[string]interface{}{
			"a": 5.5,
			"b": 3.2,
		}
		a, b, err := controller.extractNumberParams(args)
		require.NoError(t, err)
		assert.InEpsilon(t, 5.5, a, 0.0001)
		assert.InEpsilon(t, 3.2, b, 0.0001)

		// Test with int
		args = map[string]interface{}{
			"a": 5,
			"b": 3,
		}
		a, b, err = controller.extractNumberParams(args)
		require.NoError(t, err)
		assert.InEpsilon(t, 5.0, a, 0.0001)
		assert.InEpsilon(t, 3.0, b, 0.0001)

		// Test with int64
		args = map[string]interface{}{
			"a": int64(5),
			"b": int64(3),
		}
		a, b, err = controller.extractNumberParams(args)
		require.NoError(t, err)
		assert.InEpsilon(t, 5.0, a, 0.0001)
		assert.InEpsilon(t, 3.0, b, 0.0001)
	})

	t.Run("should handle invalid argument types", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		// Test with invalid arguments object
		_, _, err := controller.extractNumberParams("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "arguments must be an object")

		// Test with missing parameters
		args := map[string]interface{}{
			"a": 5.0,
			// Missing "b"
		}
		_, _, err = controller.extractNumberParams(args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "b parameter is required")

		// Test with invalid parameter type
		args = map[string]interface{}{
			"a": "invalid",
			"b": 3.0,
		}
		_, _, err = controller.extractNumberParams(args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a parameter must be a number")
	})

	t.Run("should extract calculate parameters correctly", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		args := map[string]interface{}{
			"operation": "multiply",
			"a":         6.0,
			"b":         7.0,
		}

		operation, a, b, err := controller.extractCalculateParams(args)
		require.NoError(t, err)
		assert.Equal(t, "multiply", operation)
		assert.InEpsilon(t, 6.0, a, 0.0001)
		assert.InEpsilon(t, 7.0, b, 0.0001)
	})
}

func TestMathController_RegisterWithServer(t *testing.T) {
	t.Run("should register all tools successfully", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		mockRegistrar := &MockToolRegistrar{}
		mockRegistrar.On("RegisterTool", mock.AnythingOfType("mcp.Tool"), mock.AnythingOfType("server.ToolHandlerFunc")).
			Return(nil).Times(5) // Expect 5 tools to be registered

		err := controller.RegisterWithServer(mockRegistrar)

		require.NoError(t, err)
		mockRegistrar.AssertExpectations(t)
	})

	t.Run("should handle registration error", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		mockRegistrar := &MockToolRegistrar{}
		mockRegistrar.On("RegisterTool", mock.AnythingOfType("mcp.Tool"), mock.AnythingOfType("server.ToolHandlerFunc")).
			Return(errors.New("registration failed")).Once()

		err := controller.RegisterWithServer(mockRegistrar)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to register tool")
		mockRegistrar.AssertExpectations(t)
	})
}

func TestMathController_ParameterTypeSafety(t *testing.T) {
	t.Run("should validate parameter types strictly", func(t *testing.T) {
		deps := makeMathControllerDeps()
		controller := NewMathController(deps)

		testCases := []struct {
			name      string
			value     interface{}
			expected  float64
			shouldErr bool
		}{
			{"float64", 5.5, 5.5, false},
			{"int", 5, 5.0, false},
			{"int64", int64(5), 5.0, false},
			{"string", "5", 0, true},
			{"bool", true, 0, true},
			{"nil", nil, 0, true},
			{"array", []int{1, 2, 3}, 0, true},
			{"map", map[string]int{"a": 1}, 0, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				args := map[string]interface{}{
					"test_param": tc.value,
				}

				result, err := controller.extractNumberParam(args, "test_param")

				if tc.shouldErr {
					require.Error(t, err)
					assert.Contains(t, err.Error(), "parameter must be a number")
				} else {
					require.NoError(t, err)
					assert.InEpsilon(t, tc.expected, result, 0.0001)
				}
			})
		}
	})
}
