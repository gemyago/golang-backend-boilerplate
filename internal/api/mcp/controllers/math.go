package controllers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
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

// newCalculateTool returns the MCP tool definition for generic calculations.
func (mc *MathController) newCalculateTool() mcp.Tool {
	return mcp.NewTool(
		"calculate",
		mcp.WithDescription("Perform mathematical calculations (add, subtract, multiply, divide)"),
		mcp.WithString("operation",
			mcp.Description("Mathematical operation to perform"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("a", mcp.Description("First number")),
		mcp.WithNumber("b", mcp.Description("Second number")),
	)
}

// newAddTool returns the MCP tool definition for addition.
func (mc *MathController) newAddTool() mcp.Tool {
	return mcp.NewTool(
		"add",
		mcp.WithDescription("Add two numbers together"),
		mcp.WithNumber("a", mcp.Description("First number to add")),
		mcp.WithNumber("b", mcp.Description("Second number to add")),
	)
}

// newSubtractTool returns the MCP tool definition for subtraction.
func (mc *MathController) newSubtractTool() mcp.Tool {
	return mcp.NewTool(
		"subtract",
		mcp.WithDescription("Subtract second number from first number"),
		mcp.WithNumber("a", mcp.Description("Number to subtract from")),
		mcp.WithNumber("b", mcp.Description("Number to subtract")),
	)
}

// newMultiplyTool returns the MCP tool definition for multiplication.
func (mc *MathController) newMultiplyTool() mcp.Tool {
	return mcp.NewTool(
		"multiply",
		mcp.WithDescription("Multiply two numbers together"),
		mcp.WithNumber("a", mcp.Description("First number to multiply")),
		mcp.WithNumber("b", mcp.Description("Second number to multiply")),
	)
}

// newDivideTool returns the MCP tool definition for division.
func (mc *MathController) newDivideTool() mcp.Tool {
	return mcp.NewTool(
		"divide",
		mcp.WithDescription("Divide first number by second number"),
		mcp.WithNumber("a", mcp.Description("Dividend (number to be divided)")),
		mcp.WithNumber("b", mcp.Description("Divisor (number to divide by)")),
	)
}

// handleCalculate handles the calculate tool call.
func (mc *MathController) handleCalculate(ctx context.Context,
	request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.logger.InfoContext(ctx, "Handling calculate tool call",
		slog.String("tool", request.Params.Name))

	operation, a, b, err := mc.extractCalculateParams(request.Params.Arguments)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Invalid calculate parameters", slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	mathRequest := &app.MathRequest{
		Operation: app.MathOperation(operation),
		A:         a,
		B:         b,
	}

	response, err := mc.mathService.Calculate(ctx, mathRequest)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Math calculation failed",
			slog.String("operation", operation),
			slog.Float64("a", a),
			slog.Float64("b", b),
			slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Calculation failed: %v", err)), nil
	}

	resultText := fmt.Sprintf("Result: %g (operation: %s, a: %g, b: %g)",
		response.Result, response.Operation, response.A, response.B)

	mc.logger.InfoContext(ctx, "Successfully calculated result",
		slog.Float64("result", response.Result),
		slog.String("operation", string(response.Operation)))

	return mcp.NewToolResultText(resultText), nil
}

// handleAdd handles the add tool call.
func (mc *MathController) handleAdd(ctx context.Context,
	request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.logger.InfoContext(ctx, "Handling add tool call",
		slog.String("tool", request.Params.Name))

	a, b, err := mc.extractNumberParams(request.Params.Arguments)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Invalid add parameters", slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	response, err := mc.mathService.Add(ctx, a, b)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Addition failed",
			slog.Float64("a", a),
			slog.Float64("b", b),
			slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Addition failed: %v", err)), nil
	}

	resultText := fmt.Sprintf("Result: %g + %g = %g", response.A, response.B, response.Result)

	mc.logger.InfoContext(ctx, "Successfully performed addition",
		slog.Float64("result", response.Result))

	return mcp.NewToolResultText(resultText), nil
}

// handleSubtract handles the subtract tool call.
func (mc *MathController) handleSubtract(ctx context.Context,
	request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.logger.InfoContext(ctx, "Handling subtract tool call",
		slog.String("tool", request.Params.Name))

	a, b, err := mc.extractNumberParams(request.Params.Arguments)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Invalid subtract parameters", slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	response, err := mc.mathService.Subtract(ctx, a, b)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Subtraction failed",
			slog.Float64("a", a),
			slog.Float64("b", b),
			slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Subtraction failed: %v", err)), nil
	}

	resultText := fmt.Sprintf("Result: %g - %g = %g", response.A, response.B, response.Result)

	mc.logger.InfoContext(ctx, "Successfully performed subtraction",
		slog.Float64("result", response.Result))

	return mcp.NewToolResultText(resultText), nil
}

// handleMultiply handles the multiply tool call.
func (mc *MathController) handleMultiply(ctx context.Context,
	request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.logger.InfoContext(ctx, "Handling multiply tool call",
		slog.String("tool", request.Params.Name))

	a, b, err := mc.extractNumberParams(request.Params.Arguments)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Invalid multiply parameters", slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	response, err := mc.mathService.Multiply(ctx, a, b)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Multiplication failed",
			slog.Float64("a", a),
			slog.Float64("b", b),
			slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Multiplication failed: %v", err)), nil
	}

	resultText := fmt.Sprintf("Result: %g × %g = %g", response.A, response.B, response.Result)

	mc.logger.InfoContext(ctx, "Successfully performed multiplication",
		slog.Float64("result", response.Result))

	return mcp.NewToolResultText(resultText), nil
}

// handleDivide handles the divide tool call.
func (mc *MathController) handleDivide(ctx context.Context,
	request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mc.logger.InfoContext(ctx, "Handling divide tool call",
		slog.String("tool", request.Params.Name))

	a, b, err := mc.extractNumberParams(request.Params.Arguments)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Invalid divide parameters", slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
	}

	response, err := mc.mathService.Divide(ctx, a, b)
	if err != nil {
		mc.logger.ErrorContext(ctx, "Division failed",
			slog.Float64("a", a),
			slog.Float64("b", b),
			slog.String("error", err.Error()))
		return mcp.NewToolResultError(fmt.Sprintf("Division failed: %v", err)), nil
	}

	resultText := fmt.Sprintf("Result: %g ÷ %g = %g", response.A, response.B, response.Result)

	mc.logger.InfoContext(ctx, "Successfully performed division",
		slog.Float64("result", response.Result))

	return mcp.NewToolResultText(resultText), nil
}

// extractCalculateParams extracts operation, a, and b parameters from arguments.
func (mc *MathController) extractCalculateParams(args interface{}) (string, float64, float64, error) {
	argsMap, ok := args.(map[string]interface{})
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
func (mc *MathController) extractNumberParams(args interface{}) (float64, float64, error) {
	argsMap, ok := args.(map[string]interface{})
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
	args map[string]interface{},
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

// RegisterWithServer registers all math tools with the MCP server.
func (mc *MathController) RegisterWithServer(server ToolRegistrar) error {
	tools := []struct {
		tool    mcp.Tool
		handler ToolHandler
	}{
		{mc.newCalculateTool(), mc.handleCalculate},
		{mc.newAddTool(), mc.handleAdd},
		{mc.newSubtractTool(), mc.handleSubtract},
		{mc.newMultiplyTool(), mc.handleMultiply},
		{mc.newDivideTool(), mc.handleDivide},
	}

	for _, toolInfo := range tools {
		mc.logger.Info("Registering math tool with MCP server",
			slog.String("tool_name", toolInfo.tool.Name),
			slog.String("description", toolInfo.tool.Description))

		if err := server.RegisterTool(toolInfo.tool, toolInfo.handler); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", toolInfo.tool.Name, err)
		}
	}

	mc.logger.Info("Successfully registered all math tools with MCP server",
		slog.Int("tool_count", len(tools)))

	return nil
}
