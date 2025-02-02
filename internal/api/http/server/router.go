package server

import (
	"net/http"

	"go.uber.org/dig"
)

type MuxRouterAdapter struct {
	mux *http.ServeMux
}

func (r *MuxRouterAdapter) PathValue(req *http.Request, paramName string) string {
	return req.PathValue(paramName)
}

func (r *MuxRouterAdapter) HandleRoute(method, pathPattern string, h http.Handler) {
	r.mux.Handle(method+" "+pathPattern, h)
}

func NewMuxRouterAdapter(mux *http.ServeMux) *MuxRouterAdapter {
	return &MuxRouterAdapter{
		mux: mux,
	}
}

type Group struct {
	dig.Out

	RegisterHandlers RegisterHandlersFunc `group:"server"`
}

type RegisterHandlersFunc func()

func MakeHandlersGroupFactory[TController any, THttpApp any](
	registerFunc func(TController, THttpApp),
) func(ctrl TController, app THttpApp) Group {
	return func(ctrl TController, app THttpApp) Group {
		return Group{
			RegisterHandlers: func() { registerFunc(ctrl, app) },
		}
	}
}
