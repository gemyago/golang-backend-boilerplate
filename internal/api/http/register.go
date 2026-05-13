package http //nolint:revive // Package name follows the directory structure and API domain.

import (
	"errors"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1controllers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"go.uber.org/dig"
)

// Use apigen to generate v1routes
//go:generate go run github.com/gemyago/apigen server ./v1routes.yaml ./v1routes

type V1RoutesDeps struct {
	dig.In

	*v1controllers.HealthController
	*v1controllers.EchoController
	*v1controllers.UsersController
	*v1controllers.PetsController

	RootHandler *handlers.RootHandler
}

func SetupV1Routes(deps V1RoutesDeps) { // coverage-ignore // Little value in testing wireup code.
	rootHandler := deps.RootHandler
	rootHandler.RegisterHealthRoutes(deps.HealthController)
	rootHandler.RegisterEchoRoutes(deps.EchoController)
	rootHandler.RegisterUsersRoutes(deps.UsersController)
	rootHandler.RegisterPetsRoutes(deps.PetsController)
}

func Register(container *dig.Container) error {
	return errors.Join(
		v1controllers.Register(container),
		container.Invoke(SetupV1Routes),
	)
}
