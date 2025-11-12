package petstore

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	httpservices "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
)

func makeMockDeps(t *testing.T, baseURL string) ClientDeps {
	rootLogger := diag.RootTestLogger().With("test", t.Name())
	return ClientDeps{
		ClientFactory: httpservices.NewClientFactory(httpservices.ClientFactoryDeps{
			RootLogger: rootLogger,
		}),
		RootLogger: rootLogger,
		BaseURL:    baseURL,
	}
}
