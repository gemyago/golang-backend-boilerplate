package controllers

import (
	"context"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"go.uber.org/dig"
)

// RegistryDeps contains dependencies for the controllers registry.
type RegistryDeps struct {
	dig.In

	RootLogger  *slog.Logger
	TimeService *app.TimeService
}

// Registry manages all MCP controllers and their registration.
type Registry struct {
	logger         *slog.Logger
	timeController *TimeController
}

// NewControllersRegistry creates a new controllers registry.
func NewControllersRegistry(deps RegistryDeps) *Registry {
	// Create individual controllers
	timeController := NewTimeController(TimeControllerDeps{
		RootLogger:  deps.RootLogger,
		TimeService: deps.TimeService,
	})

	return &Registry{
		logger:         deps.RootLogger.WithGroup("mcp.controllers-registry"),
		timeController: timeController,
	}
}

// RegisterAllControllers registers all available MCP controllers with the given server.
func (cr *Registry) RegisterAllControllers(ctx context.Context, server ToolRegistrar) error {
	cr.logger.InfoContext(ctx, "Registering all MCP controllers")

	// Register time controller
	if err := cr.timeController.RegisterWithServer(server); err != nil {
		cr.logger.ErrorContext(ctx, "Failed to register time controller",
			slog.String("error", err.Error()))
		return err
	}

	// TODO: Add math controller registration when implemented
	// if err := cr.mathController.RegisterWithServer(server); err != nil {
	//     return err
	// }

	cr.logger.InfoContext(ctx, "Successfully registered all MCP controllers")
	return nil
}

// GetTimeController returns the time controller instance.
func (cr *Registry) GetTimeController() *TimeController {
	return cr.timeController
}

// TODO: Add GetMathController when implemented
// func (cr *Registry) GetMathController() *MathController {
//     return cr.mathController
// }
