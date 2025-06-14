package controllers

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/dig"
)

// ControllersRegistryDeps contains dependencies for the controllers registry.
type ControllersRegistryDeps struct {
	dig.In

	RootLogger     *slog.Logger
	MathController *MathController
	TimeController *TimeController
}

// ControllersRegistry manages all MCP controllers.
type ControllersRegistry struct {
	logger         *slog.Logger
	mathController *MathController
	timeController *TimeController
}

// NewControllersRegistry creates a new controllers registry.
func NewControllersRegistry(deps ControllersRegistryDeps) *ControllersRegistry {
	return &ControllersRegistry{
		logger:         deps.RootLogger.WithGroup("mcp.controllers-registry"),
		mathController: deps.MathController,
		timeController: deps.TimeController,
	}
}

// RegisterAllControllers registers all controllers with the MCP server.
func (r *ControllersRegistry) RegisterAllControllers(server interface {
	AddTools(tools ...server.ServerTool)
}) error {
	r.logger.Info("Registering all MCP controllers")

	// Register math controller
	if err := r.mathController.RegisterWithServer(server); err != nil {
		return err
	}

	// Register time controller
	if err := r.timeController.RegisterWithServer(server); err != nil {
		return err
	}

	r.logger.Info("Successfully registered all MCP controllers")
	return nil
}

// GetMathController returns the math controller instance.
func (r *ControllersRegistry) GetMathController() *MathController {
	return r.mathController
}

// GetTimeController returns the time controller instance.
func (r *ControllersRegistry) GetTimeController() *TimeController {
	return r.timeController
}
