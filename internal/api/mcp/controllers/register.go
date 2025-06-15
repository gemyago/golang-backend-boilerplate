package controllers

import (
	"github.com/gemyago/golang-backend-boilerplate/internal/api/mcp/server"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

type controllerResult struct {
	dig.Out

	Controller server.ToolController `group:"mcp-controllers"`
}

func newControllerResult[T server.ToolController](controller T) controllerResult {
	return controllerResult{
		Controller: controller,
	}
}

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		NewMathController,
		NewTimeController,
		newControllerResult[*MathController],
		newControllerResult[*TimeController],
	)
}
