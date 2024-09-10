package routes

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/pkg/diag"
)

type router interface {
	Handle(pattern string, handler http.Handler)
}

func WriteData(req *http.Request, log *slog.Logger, writer io.Writer, data []byte) {
	if _, err := writer.Write(data); err != nil {
		log.ErrorContext(req.Context(), "Failed to write response", diag.ErrAttr(err))
	}
}
