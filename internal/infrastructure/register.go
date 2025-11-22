package infrastructure

import (
	"context"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

func Register(rootCtx context.Context, container *dig.Container) error {
	return di.ProvideAll(container,
		NewTimeProvider,
		di.ProvideValue(time.NewTicker),
		NewShutdownHooks,

		// We can't directly use shutdown hooks in diag, since diag is used everywhere.
		// So we need to register the implementation somewhere.
		// We do it here until we have a better way.
		di.ProvideImplementation[*ShutdownHooks, diag.ShutdownHooks],

		httpservices.NewClientFactory,
		newDBProvider(rootCtx),
		di.ProvideFactoryAs[app.UsersRepository](newUsersRepository),
		di.ProvideFactoryAs[app.PetsRepository](newPetsRepository),
		di.ProvideFactoryAs[app.PetstoreClient](func(deps petstore.ClientDeps) *petstore.Client {
			return petstore.NewClient(deps)
		}),
	)
}
