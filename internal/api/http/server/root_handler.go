package server

import (
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/middleware"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
)

func NewRootHandler(
	rootLogger *slog.Logger,
) *handlers.RootHandler { // coverage-ignore // Little value in testing wireup code.
	logger := rootLogger.WithGroup("http")

	rootHandler := handlers.NewRootHandler(
		(*HTTPRouter)(http.NewServeMux()),
		handlers.WithLogger(logger),
		handlers.WithActionErrorHandler(
			middleware.NewAppErrorHandler(rootLogger),
		),
	)

	return rootHandler
}
