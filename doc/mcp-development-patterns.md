# MCP Tool Development Patterns

This document outlines the development patterns and best practices for creating new MCP tools in the golang-backend-boilerplate project.

## Architecture Overview

The MCP tools extension follows a clean architecture pattern with clear separation of concerns:

```
cmd/mcp/                    # CLI commands and entry points
├── main.go                 # Main entry point
├── root.go                 # Root command configuration
├── stdio.go                # stdio transport handler
└── http.go                 # HTTP transport handler

internal/api/mcp/           # MCP protocol layer
├── server/                 # MCP server implementation
│   └── server.go          # Server initialization and management
└── controllers/           # MCP tool controllers
    ├── register.go        # Tool registration
    ├── time.go           # Time tool controller
    └── math.go           # Math tool controller

internal/app/              # Business logic layer
├── register.go           # Service registration
├── time.go              # Time service
└── math.go              # Math service
```

## Development Workflow

### 1. Create the Service Layer (internal/app/)

Start by implementing the business logic in the app layer.

**Example: Creating a new calculator service**

```go
package app

import (
    "context"
    "fmt"
    "log/slog"
    "go.uber.org/dig"
)

// CalculatorOperation represents supported calculator operations
type CalculatorOperation string

const (
    CalculatorOperationPower     CalculatorOperation = "power"
    CalculatorOperationSquareRoot CalculatorOperation = "sqrt"
)

// CalculatorRequest represents a calculator operation request
type CalculatorRequest struct {
    Operation CalculatorOperation `json:"operation"`
    Base      float64             `json:"base"`
    Exponent  float64             `json:"exponent,omitempty"`
}

// CalculatorResponse represents a calculator operation response
type CalculatorResponse struct {
    Result    float64             `json:"result"`
    Operation CalculatorOperation `json:"operation"`
    Input     interface{}         `json:"input"`
}

// CalculatorServiceDeps contains dependencies for the calculator service
type CalculatorServiceDeps struct {
    dig.In
    RootLogger *slog.Logger
}

// CalculatorService provides advanced mathematical operations
type CalculatorService struct {
    logger *slog.Logger
}

// Power calculates base raised to the power of exponent
func (svc *CalculatorService) Power(ctx context.Context, base, exponent float64) (*CalculatorResponse, error) {
    svc.logger.InfoContext(ctx, "Performing power operation",
        slog.Float64("base", base),
        slog.Float64("exponent", exponent))

    // Implementation logic here
    result := math.Pow(base, exponent)

    response := &CalculatorResponse{
        Result:    result,
        Operation: CalculatorOperationPower,
        Input:     map[string]float64{"base": base, "exponent": exponent},
    }

    svc.logger.InfoContext(ctx, "Power operation completed",
        slog.Float64("result", result))

    return response, nil
}

// NewCalculatorService creates a new calculator service instance
func NewCalculatorService(deps CalculatorServiceDeps) *CalculatorService {
    return &CalculatorService{
        logger: deps.RootLogger.WithGroup("app.calculator-service"),
    }
}
```

### 2. Write Service Tests

Follow TDD principles and create comprehensive tests.

**Example: Calculator service tests**

```go
package app

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestCalculatorService_Power(t *testing.T) {
    t.Run("should calculate power correctly", func(t *testing.T) {
        // Given
        deps := makeMockDeps(t)
        service := NewCalculatorService(deps)
        ctx := context.Background()

        // When
        result, err := service.Power(ctx, 2, 3)

        // Then
        require.NoError(t, err)
        require.NotNil(t, result)
        require.Equal(t, 8.0, result.Result)
        require.Equal(t, CalculatorOperationPower, result.Operation)
    })

    t.Run("should handle edge cases", func(t *testing.T) {
        cases := []struct {
            name     string
            base     float64
            exponent float64
            expected float64
        }{
            {"zero power", 5, 0, 1},
            {"power of one", 7, 1, 7},
            {"negative base", -2, 2, 4},
        }

        deps := makeMockDeps(t)
        service := NewCalculatorService(deps)
        ctx := context.Background()

        for _, tc := range cases {
            t.Run(tc.name, func(t *testing.T) {
                result, err := service.Power(ctx, tc.base, tc.exponent)
                require.NoError(t, err)
                require.Equal(t, tc.expected, result.Result)
            })
        }
    })
}
```

### 3. Create MCP Controller

Implement the MCP protocol layer that bridges the service to MCP tools.

**Example: Calculator MCP controller**

```go
package controllers

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/gemyago/golang-backend-boilerplate/internal/app"
    "github.com/mark3labs/mcp-go/mcp"
    "go.uber.org/dig"
)

// CalculatorControllerDeps contains dependencies for the calculator MCP controller
type CalculatorControllerDeps struct {
    dig.In
    CalculatorService *app.CalculatorService
    RootLogger        *slog.Logger
}

// CalculatorController implements MCP tools for calculator operations
type CalculatorController struct {
    calculatorService *app.CalculatorService
    logger            *slog.Logger
}

// NewCalculatorController creates a new calculator MCP controller
func NewCalculatorController(deps CalculatorControllerDeps) *CalculatorController {
    return &CalculatorController{
        calculatorService: deps.CalculatorService,
        logger:            deps.RootLogger.WithGroup("mcp.calculator-controller"),
    }
}

// GetPowerTool returns the MCP tool definition for power calculation
func (cc *CalculatorController) GetPowerTool() mcp.Tool {
    return mcp.NewTool(
        "power",
        mcp.WithDescription("Calculate base raised to the power of exponent"),
        mcp.WithNumber("base", mcp.Description("Base number")),
        mcp.WithNumber("exponent", mcp.Description("Exponent")),
    )
}

// HandlePower handles the power tool call
func (cc *CalculatorController) HandlePower(ctx context.Context,
    request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    cc.logger.InfoContext(ctx, "Handling power tool call",
        slog.String("tool", request.Params.Name))

    base, exponent, err := cc.extractPowerParams(request.Params.Arguments)
    if err != nil {
        cc.logger.ErrorContext(ctx, "Invalid power parameters", slog.String("error", err.Error()))
        return mcp.NewToolResultError(fmt.Sprintf("Invalid parameters: %v", err)), nil
    }

    response, err := cc.calculatorService.Power(ctx, base, exponent)
    if err != nil {
        cc.logger.ErrorContext(ctx, "Power calculation failed",
            slog.Float64("base", base),
            slog.Float64("exponent", exponent),
            slog.String("error", err.Error()))
        return mcp.NewToolResultError(fmt.Sprintf("Power calculation failed: %v", err)), nil
    }

    resultText := fmt.Sprintf("Result: %g^%g = %g", base, exponent, response.Result)

    cc.logger.InfoContext(ctx, "Successfully calculated power",
        slog.Float64("result", response.Result))

    return mcp.NewToolResultText(resultText), nil
}

// extractPowerParams extracts and validates power parameters
func (cc *CalculatorController) extractPowerParams(args interface{}) (float64, float64, error) {
    argsMap, ok := args.(map[string]interface{})
    if !ok {
        return 0, 0, fmt.Errorf("arguments must be an object")
    }

    base, err := cc.extractNumberParam(argsMap, "base")
    if err != nil {
        return 0, 0, err
    }

    exponent, err := cc.extractNumberParam(argsMap, "exponent")
    if err != nil {
        return 0, 0, err
    }

    return base, exponent, nil
}

// extractNumberParam extracts and validates a number parameter
func (cc *CalculatorController) extractNumberParam(
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

// RegisterWithServer registers all calculator tools with the MCP server
func (cc *CalculatorController) RegisterWithServer(server ToolRegistrar) error {
    tools := []struct {
        tool    mcp.Tool
        handler ToolHandler
    }{
        {cc.GetPowerTool(), cc.HandlePower},
    }

    for _, toolInfo := range tools {
        cc.logger.Info("Registering calculator tool with MCP server",
            slog.String("tool_name", toolInfo.tool.Name),
            slog.String("description", toolInfo.tool.Description))

        if err := server.RegisterTool(toolInfo.tool, toolInfo.handler); err != nil {
            return fmt.Errorf("failed to register tool %s: %w", toolInfo.tool.Name, err)
        }
    }

    cc.logger.Info("Successfully registered all calculator tools with MCP server",
        slog.Int("tool_count", len(tools)))

    return nil
}
```

### 4. Write Controller Tests

Test the MCP protocol layer including parameter validation and error handling.

**Example: Calculator controller tests**

```go
package controllers

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/mark3labs/mcp-go/mcp"
)

func TestCalculatorController_ToolDefinitions(t *testing.T) {
    t.Run("should return power tool with correct schema", func(t *testing.T) {
        // Given
        deps := makeMockCalculatorControllerDeps(t)
        controller := NewCalculatorController(deps)

        // When
        tool := controller.GetPowerTool()

        // Then
        require.Equal(t, "power", tool.Name)
        require.Equal(t, "Calculate base raised to the power of exponent", tool.Description)
        require.Contains(t, tool.InputSchema.Properties, "base")
        require.Contains(t, tool.InputSchema.Properties, "exponent")
    })
}

func TestCalculatorController_HandlePower(t *testing.T) {
    t.Run("should handle power calculation successfully", func(t *testing.T) {
        // Given
        deps := makeMockCalculatorControllerDeps(t)
        controller := NewCalculatorController(deps)
        ctx := context.Background()

        request := mcp.CallToolRequest{
            Params: mcp.CallToolRequestParams{
                Name: "power",
                Arguments: map[string]interface{}{
                    "base":     2.0,
                    "exponent": 3.0,
                },
            },
        }

        // When
        result, err := controller.HandlePower(ctx, request)

        // Then
        require.NoError(t, err)
        require.NotNil(t, result)
        require.False(t, result.IsError)
    })

    t.Run("should handle invalid parameters", func(t *testing.T) {
        // Given
        deps := makeMockCalculatorControllerDeps(t)
        controller := NewCalculatorController(deps)
        ctx := context.Background()

        request := mcp.CallToolRequest{
            Params: mcp.CallToolRequestParams{
                Name: "power",
                Arguments: map[string]interface{}{
                    "base": "not_a_number",
                },
            },
        }

        // When
        result, err := controller.HandlePower(ctx, request)

        // Then
        require.NoError(t, err)
        require.True(t, result.IsError)
    })
}
```

### 5. Register Services and Controllers

Update the registration files to include your new services and controllers.

**Update internal/app/register.go:**

```go
func Register(container *dig.Container) error {
    constructors := []interface{}{
        // ... existing constructors
        NewCalculatorService,
    }
    
    // ... registration logic
}
```

**Update internal/api/mcp/controllers/register.go:**

```go
type RegistryDeps struct {
    dig.In
    // ... existing dependencies
    CalculatorService *app.CalculatorService
}

type Registry struct {
    // ... existing controllers
    calculatorController *CalculatorController
}

func NewRegistry(deps RegistryDeps) *Registry {
    return &Registry{
        // ... existing controllers
        calculatorController: NewCalculatorController(CalculatorControllerDeps{
            CalculatorService: deps.CalculatorService,
            RootLogger:        deps.RootLogger,
        }),
    }
}

func (r *Registry) RegisterAllControllers(server ToolRegistrar) error {
    controllers := []ControllerRegistrar{
        // ... existing controllers
        r.calculatorController,
    }
    
    // ... registration logic
}
```

## Design Patterns

### 1. Dependency Injection Pattern

Use dig for dependency injection to maintain loose coupling:

```go
type ServiceDeps struct {
    dig.In
    RootLogger *slog.Logger
    // Add other dependencies as needed
}

func NewService(deps ServiceDeps) *Service {
    return &Service{
        logger: deps.RootLogger.WithGroup("service-name"),
    }
}
```

### 2. Error Handling Pattern

Always return structured errors with context:

```go
func (svc *Service) Operation(ctx context.Context, input Input) (*Response, error) {
    svc.logger.InfoContext(ctx, "Starting operation", slog.Any("input", input))
    
    if err := validateInput(input); err != nil {
        svc.logger.ErrorContext(ctx, "Invalid input", slog.String("error", err.Error()))
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    result, err := doWork(input)
    if err != nil {
        svc.logger.ErrorContext(ctx, "Operation failed", slog.String("error", err.Error()))
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    svc.logger.InfoContext(ctx, "Operation completed", slog.Any("result", result))
    return result, nil
}
```

### 3. MCP Tool Definition Pattern

Define tools with comprehensive schemas:

```go
func (c *Controller) GetToolName() mcp.Tool {
    return mcp.NewTool(
        "tool_name",
        mcp.WithDescription("Clear, concise description of what the tool does"),
        mcp.WithString("param1", 
            mcp.Description("Parameter description"),
            mcp.Enum("option1", "option2"), // for enumerated values
        ),
        mcp.WithNumber("param2", mcp.Description("Numeric parameter description")),
        mcp.WithBoolean("param3", mcp.Description("Boolean parameter description")),
    )
}
```

### 4. Parameter Validation Pattern

Create reusable parameter extraction functions:

```go
func (c *Controller) extractCommonParams(args interface{}) (CommonParams, error) {
    argsMap, ok := args.(map[string]interface{})
    if !ok {
        return CommonParams{}, fmt.Errorf("arguments must be an object")
    }

    var params CommonParams
    var err error
    
    params.StringParam, err = c.extractStringParam(argsMap, "string_param")
    if err != nil {
        return CommonParams{}, err
    }
    
    params.NumberParam, err = c.extractNumberParam(argsMap, "number_param")
    if err != nil {
        return CommonParams{}, err
    }
    
    return params, nil
}
```

### 5. Logging Pattern

Use structured logging consistently:

```go
func (svc *Service) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    svc.logger.InfoContext(ctx, "Processing request",
        slog.String("request_id", req.ID),
        slog.String("operation", req.Operation))
    
    // ... processing logic
    
    svc.logger.InfoContext(ctx, "Request processed successfully",
        slog.String("request_id", req.ID),
        slog.Any("response", response))
    
    return response, nil
}
```

## Testing Patterns

### 1. Test Structure Pattern

Use table-driven tests for comprehensive coverage:

```go
func TestService_Operation(t *testing.T) {
    testCases := []struct {
        name        string
        input       Input
        expectError bool
        expectedResult *Response
    }{
        {
            name: "valid input",
            input: Input{Value: 10},
            expectError: false,
            expectedResult: &Response{Result: 20},
        },
        {
            name: "invalid input",
            input: Input{Value: -1},
            expectError: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Given
            deps := makeMockDeps(t)
            service := NewService(deps)
            ctx := context.Background()

            // When
            result, err := service.Operation(ctx, tc.input)

            // Then
            if tc.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.expectedResult, result)
            }
        })
    }
}
```

### 2. Mock Setup Pattern

Create reusable mock setup functions:

```go
func makeMockDeps(t *testing.T) ServiceDeps {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    
    return ServiceDeps{
        RootLogger: logger,
    }
}

func makeMockControllerDeps(t *testing.T) ControllerDeps {
    mockService := &MockService{}
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    
    return ControllerDeps{
        Service:    mockService,
        RootLogger: logger,
    }
}
```

### 3. Integration Test Pattern

Test the complete flow from MCP request to response:

```go
func TestController_Integration(t *testing.T) {
    t.Run("should handle complete request flow", func(t *testing.T) {
        // Given - Real dependencies, not mocks
        deps := makeRealDeps(t)
        controller := NewController(deps)
        ctx := context.Background()

        request := mcp.CallToolRequest{
            Params: mcp.CallToolRequestParams{
                Name: "tool_name",
                Arguments: map[string]interface{}{
                    "param": "value",
                },
            },
        }

        // When
        result, err := controller.HandleTool(ctx, request)

        // Then
        require.NoError(t, err)
        require.NotNil(t, result)
        require.False(t, result.IsError)
        
        // Verify response content
        require.Len(t, result.Content, 1)
        require.Equal(t, "text", result.Content[0].Type)
        require.Contains(t, result.Content[0].Text, "expected content")
    })
}
```

## Performance Patterns

### 1. Context Management

Always respect context cancellation:

```go
func (svc *Service) LongRunningOperation(ctx context.Context) error {
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Continue processing
        }
        
        // Do work...
        time.Sleep(10 * time.Millisecond)
    }
    return nil
}
```

### 2. Resource Management

Use proper resource cleanup:

```go
func (svc *Service) ProcessWithResource(ctx context.Context) error {
    resource, err := acquireResource()
    if err != nil {
        return err
    }
    defer resource.Close()
    
    // Use resource...
    return nil
}
```

## Security Patterns

### 1. Input Validation

Always validate and sanitize inputs:

```go
func validateInput(input *Input) error {
    if input == nil {
        return errors.New("input cannot be nil")
    }
    
    if input.Value < 0 || input.Value > 1000 {
        return errors.New("value must be between 0 and 1000")
    }
    
    if strings.TrimSpace(input.Name) == "" {
        return errors.New("name cannot be empty")
    }
    
    return nil
}
```

### 2. Error Information Disclosure

Don't leak sensitive information in error messages:

```go
func (svc *Service) ProcessSensitiveData(ctx context.Context, data *SensitiveData) error {
    if err := validateSensitiveData(data); err != nil {
        svc.logger.ErrorContext(ctx, "Validation failed", 
            slog.String("error", err.Error()),
            slog.String("data_id", data.ID)) // Log details for debugging
        
        // Return generic error to client
        return errors.New("invalid input data")
    }
    
    // ... processing
    return nil
}
```

## Testing Commands

Run tests with appropriate flags:

```bash
# Run unit tests for specific package
go test -v ./internal/app -run TestServiceName

# Run with coverage
go test -v -cover ./internal/app

# Run integration tests
go test -v ./internal/api/mcp/controllers -run TestIntegration

# Run all tests
make test

# Watch mode for development
gow test -v ./internal/app -run TestServiceName
```

## Code Quality

### 1. Follow Go Conventions

- Use meaningful names
- Keep functions small and focused
- Document exported functions and types
- Handle errors explicitly

### 2. Use Linters

Ensure code quality with linters:

```bash
golangci-lint run
```

### 3. Test Coverage

Maintain high test coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Documentation

### 1. Code Documentation

Document public APIs:

```go
// CalculatorService provides advanced mathematical operations.
// It supports power calculations, square roots, and other complex operations.
type CalculatorService struct {
    logger *slog.Logger
}

// Power calculates base raised to the power of exponent.
// It returns an error if the calculation would result in overflow or invalid results.
func (svc *CalculatorService) Power(ctx context.Context, base, exponent float64) (*CalculatorResponse, error) {
    // Implementation...
}
```

### 2. API Documentation

Update usage documentation when adding new tools:

- Add tool descriptions to usage examples
- Include parameter specifications
- Provide integration examples
- Document error scenarios

By following these patterns, you'll create maintainable, testable, and robust MCP tools that integrate seamlessly with the existing architecture. 