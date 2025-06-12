package controllers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"
)

// TimeControllerDeps contains dependencies for the time MCP controller.
type TimeControllerDeps struct {
	dig.In

	RootLogger  *slog.Logger
	TimeService *app.TimeService
}

// TimeController provides MCP time tool functionality.
type TimeController struct {
	logger      *slog.Logger
	timeService *app.TimeService
}

// NewTimeController creates a new time MCP controller.
func NewTimeController(deps TimeControllerDeps) *TimeController {
	return &TimeController{
		logger:      deps.RootLogger.WithGroup("mcp.time-controller"),
		timeService: deps.TimeService,
	}
}

// GetTimeTool returns the MCP tool definition for getting current time.
func (tc *TimeController) GetTimeTool() mcp.Tool {
	return mcp.NewTool(
		"get_current_time",
		mcp.WithDescription("Get the current date and time in various formats"),
		mcp.WithString("format",
			mcp.Description("Time format to return (iso, rfc3339, or unix)"),
			mcp.Enum("iso", "rfc3339", "unix"),
		),
	)
}

// HandleGetCurrentTime handles MCP tool calls for getting current time.
func (tc *TimeController) HandleGetCurrentTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tc.logger.InfoContext(ctx, "Handling get_current_time tool call",
		slog.String("tool", request.Params.Name))

	// Extract format parameter using mcp helper
	format := mcp.ParseString(request, "format", "iso")

	// Convert to app layer format
	var appFormat app.TimeFormat
	switch format {
	case "iso":
		appFormat = app.TimeFormatISO
	case "rfc3339":
		appFormat = app.TimeFormatRFC3339
	case "unix":
		appFormat = app.TimeFormatUnix
	default:
		appFormat = app.TimeFormatISO // Default to ISO
	}

	// Call the time service
	timeReq := &app.TimeRequest{
		Format: appFormat,
	}

	timeResponse, err := tc.timeService.GetCurrentTime(ctx, timeReq)
	if err != nil {
		tc.logger.ErrorContext(ctx, "Failed to get current time",
			slog.String("error", err.Error()),
			slog.String("format", format))
		return mcp.NewToolResultErrorFromErr("Failed to get current time", err), nil
	}

	tc.logger.InfoContext(ctx, "Successfully retrieved current time",
		slog.String("time", timeResponse.Time),
		slog.String("format", timeResponse.Format))

	// Create MCP tool result using helper
	resultText := fmt.Sprintf("Current time: %s (format: %s)", timeResponse.Time, timeResponse.Format)
	return mcp.NewToolResultText(resultText), nil
}

// RegisterWithServer registers the time tool with the MCP server.
func (tc *TimeController) RegisterWithServer(server ToolRegistrar) error {
	tool := tc.GetTimeTool()

	tc.logger.Info("Registering time tool with MCP server",
		slog.String("tool_name", tool.Name),
		slog.String("description", tool.Description))

	return server.RegisterTool(tool, tc.HandleGetCurrentTime)
}

// ToolRegistrar interface for registering tools with MCP server
// This allows us to decouple the controller from the specific server implementation
type ToolRegistrar interface {
	RegisterTool(tool mcp.Tool, handler ToolHandler) error
}

// Import the ToolHandler type alias from the server package
type ToolHandler = server.ToolHandlerFunc
