package v1controllers

import (
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"go.uber.org/dig"
)

func Register(container *dig.Container) error {
	return di.ProvideAll(container,
		newEchoController,
		newUsersController,
		newPetsController,
		di.ProvideValue(&HealthController{}),
		di.ProvideValue(&UsersMapper{}),

		// Application layer implementations
		di.ProvideImplementation[*app.UserCommands, UserCommands],
		di.ProvideImplementation[*app.UserQueries, UserQueries],
		di.ProvideImplementation[*app.PetsCommands, PetsCommands],
		di.ProvideImplementation[*app.PetsQueries, PetsQueries],
	)
}
