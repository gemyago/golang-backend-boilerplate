package healthcheckv1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/ping", handlePing)
	return r
}
