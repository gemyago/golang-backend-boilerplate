package otel

import (
	"net/http"
)

type HTTPMiddlewareFactory func(
	operation string,
) func(http.Handler) http.Handler

func NewNoopOTELMiddlewareFactory() HTTPMiddlewareFactory {
	return func(
		string,
	) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
}
