package infrastructure

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

func Register(rootCtx context.Context, container *dig.Container) error {
	return di.ProvideAll(container,
		httpclient.NewClientFactory,
		newDBProvider(rootCtx),
		di.ProvideImplementation[*Database, app.Queryer],
		di.ProvideFactoryAs[app.UsersRepository](NewUsersRepository),
		di.ProvideFactoryAs[app.PetsRepository](newPetsRepository),
		di.ProvideFactoryAs[app.PetstoreClient](func(deps petstore.ClientDeps) *petstore.Client {
			return petstore.NewClient(deps)
		}),
	)
}
