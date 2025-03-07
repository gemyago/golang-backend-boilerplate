package app

import (
	"context"
	"log/slog"

	"github.com/gemyago/golang-backend-boilerplate/internal/app/models"
	"go.uber.org/dig"
)

// Minimalistic application layer service example.

type EchoServiceDeps struct {
	dig.In

	RootLogger *slog.Logger
}

type EchoService struct {
	logger *slog.Logger
}

func (svc *EchoService) SendEcho(
	ctx context.Context,
	data *models.SendEchoParams,
) (*models.EchoResponsePayload, error) {
	svc.logger.InfoContext(ctx, "Going to echo data", slog.String("message", data.Payload.Message))
	return &models.EchoResponsePayload{
		Message: data.Payload.Message,
	}, nil
}

func NewEchoService(deps EchoServiceDeps) *EchoService {
	return &EchoService{
		logger: deps.RootLogger.WithGroup("app.echo-service"),
	}
}
