package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
)

func NewAppErrorHandler(rootLogger *slog.Logger) func(w http.ResponseWriter, r *http.Request, err error) {
	logger := rootLogger.WithGroup("error-handler")
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var errNotFound *app.NotFoundError
		var errInvalidInput *app.InvalidInputError
		var errConflict *app.ConflictError
		logLevel := slog.LevelWarn
		switch {
		case errors.As(err, &errInvalidInput):
			w.WriteHeader(http.StatusBadRequest)
		case errors.As(err, &errConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.As(err, &errNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			logLevel = slog.LevelError
			w.WriteHeader(http.StatusInternalServerError)
		}
		logger.Log(r.Context(), logLevel, "Failed to process request", telemetry.ErrAttr(err))
	}
}
