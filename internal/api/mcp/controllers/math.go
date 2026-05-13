package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"
)

// MathController provides MCP tools for mathematical operations.

// MathControllerDeps contains dependencies for the math MCP controller.
type MathControllerDeps struct {
	dig.In

	MathService *app.MathService
	RootLogger  *slog.Logger
}

// MathController implements MCP tools for mathematical operations.
type MathController struct {
	mathService *app.MathService
	logger      *slog.Logger
}

// NewMathController creates a new math MCP controller.
func NewMathController(deps MathControllerDeps) *MathController {
	return &MathController{
		mathService: deps.MathService,
		logger:      deps.RootLogger.WithGroup("mcp.math-controller"),
	}
}

// newCalculateServerTool returns a server tool for generic calculations.
func (mc *MathController) newCalculateServerTool() mcpserver.ServerTool {
	tool := mcp.NewTool(
		"calculate",
		mcp.WithDescription("Perform mathematical calculations (add, subtract, multiply, divide)"),
		mcp.WithString("operation",
			mcp.Description("Mathematical operation to perform"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("a", mcp.Description("First number")),
		mcp.WithNumber("b", mcp.Description("Second number")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		operation, a, b, err := mc.extractCalculateParams(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
		}

		mathRequest := &app.MathRequest{
			Operation: app.MathOperation(operation),
			A:         a,
			B:         b,
		}

		response, err := mc.mathService.Calculate(ctx, mathRequest)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Calculation failed: %v", err)), nil
		}

		resultText := fmt.Sprintf("Result: %g (operation: %s, a: %g, b: %g)",
			response.Result, response.Operation, response.A, response.B)

		return mcp.NewToolResultText(resultText), nil
	}

	return mcpserver.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

func (mc *MathController) newBinaryOpServerTool(
	name string,
	description string,
	opFunc func(ctx context.Context, a, b float64) (*app.MathResponse, error),
	resultFmt string,
) mcpserver.ServerTool {
	tool := mcp.NewTool(
		name,
		mcp.WithDescription(description),
		mcp.WithNumber("a", mcp.Description("First number")),
		mcp.WithNumber("b", mcp.Description("Second number")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		a, b, err := mc.extractNumberParams(request.Params.Arguments)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
		}

		response, err := opFunc(ctx, a, b)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("%s failed: %v", name, err)), nil
		}

		resultText := fmt.Sprintf(resultFmt, response.A, response.B, response.Result)
		return mcp.NewToolResultText(resultText), nil
	}

	return mcpserver.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

// newAddServerTool returns a server tool for addition.
func (mc *MathController) newAddServerTool() mcpserver.ServerTool {
	return mc.newBinaryOpServerTool(
		"add",
		"Add two numbers together",
		mc.mathService.Add,
		"Result: %g + %g = %g",
	)
}

// newSubtractServerTool returns a server tool for subtraction.
func (mc *MathController) newSubtractServerTool() mcpserver.ServerTool {
	return mc.newBinaryOpServerTool(
		"subtract",
		"Subtract second number from first number",
		mc.mathService.Subtract,
		"Result: %g - %g = %g",
	)
}

// newMultiplyServerTool returns a server tool for multiplication.
func (mc *MathController) newMultiplyServerTool() mcpserver.ServerTool {
	return mc.newBinaryOpServerTool(
		"multiply",
		"Multiply two numbers together",
		mc.mathService.Multiply,
		"Result: %g × %g = %g",
	)
}

// newDivideServerTool returns a server tool for division.
func (mc *MathController) newDivideServerTool() mcpserver.ServerTool {
	return mc.newBinaryOpServerTool(
		"divide",
		"Divide first number by second number",
		mc.mathService.Divide,
		"Result: %g ÷ %g = %g",
	)
}

// extractCalculateParams extracts operation, a, and b parameters from arguments.
func (mc *MathController) extractCalculateParams(args any) (string, float64, float64, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return "", 0, 0, errors.New("arguments must be an object")
	}

	operation, ok := argsMap["operation"].(string)
	if !ok {
		return "", 0, 0, errors.New("operation parameter is required and must be a string")
	}

	a, err := mc.extractNumberParam(argsMap, "a")
	if err != nil {
		return "", 0, 0, err
	}

	b, err := mc.extractNumberParam(argsMap, "b")
	if err != nil {
		return "", 0, 0, err
	}

	return operation, a, b, nil
}

// extractNumberParams extracts a and b number parameters from arguments.
func (mc *MathController) extractNumberParams(args any) (float64, float64, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return 0, 0, errors.New("arguments must be an object")
	}

	a, err := mc.extractNumberParam(argsMap, "a")
	if err != nil {
		return 0, 0, err
	}

	b, err := mc.extractNumberParam(argsMap, "b")
	if err != nil {
		return 0, 0, err
	}

	return a, b, nil
}

// extractNumberParam extracts and validates a number parameter from args.
func (mc *MathController) extractNumberParam(
	args map[string]any,
	paramName string,
) (float64, error) {
	value, exists := args[paramName]
	if !exists {
		return 0, fmt.Errorf("%s parameter is required", paramName)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("%s parameter must be a number", paramName)
	}
}

// NewTools returns all math tools.
// Satisfies the ToolsFactory interface.
func (mc *MathController) NewTools() []mcpserver.ServerTool {
	return []mcpserver.ServerTool{
		mc.newCalculateServerTool(),
		mc.newAddServerTool(),
		mc.newSubtractServerTool(),
		mc.newMultiplyServerTool(),
		mc.newDivideServerTool(),
	}
}
