package services

import (
	"context"
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/di"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/http"
	"go.uber.org/dig"
)

func Register(rootCtx context.Context, container *dig.Container) error {
	return di.ProvideAll(container,
		NewTimeProvider,
		di.ProvideValue(time.NewTicker),
		NewShutdownHooks,
		httpservices.NewClientFactory,
		newDBProvider(rootCtx),
	)
}
