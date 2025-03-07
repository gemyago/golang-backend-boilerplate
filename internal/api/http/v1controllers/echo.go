package v1controllers

import (
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/app/models"
)

type EchoController struct {
	*app.EchoService
}

func (c EchoController) SendEcho(b handlers.HandlerBuilder[
	*models.SendEchoParams,
	*models.EchoResponsePayload,
]) http.Handler {
	return b.HandleWith(c.EchoService.SendEcho)
}

var _ handlers.EchoController = (*EchoController)(nil)

func newEchoController(echoService *app.EchoService) *EchoController {
	return &EchoController{EchoService: echoService}
}
