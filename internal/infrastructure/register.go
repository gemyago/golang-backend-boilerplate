package infrastructure

import (
	"context"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/petstore"
	"go.uber.org/dig"
)

func Register(rootCtx context.Context, container *dig.Container) error {
	return di.ProvideAll(container,
		httpservices.NewClientFactory,
		newDBProvider(rootCtx),
		di.ProvideFactoryAs[app.UsersRepository](newUsersRepository),
		di.ProvideFactoryAs[app.PetsRepository](newPetsRepository),
		di.ProvideFactoryAs[app.PetstoreClient](func(deps petstore.ClientDeps) *petstore.Client {
			return petstore.NewClient(deps)
		}),
	)
}
