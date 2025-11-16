package v1controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/handlers"
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"go.uber.org/dig"
)

type UsersController struct {
	commands UserCommands
	queries  UserQueries
}

type UsersControllerDeps struct {
	dig.In

	UserCommands
	UserQueries

	RootLogger *slog.Logger
}

func newUsersController(deps UsersControllerDeps) *UsersController {
	return &UsersController{
		commands: deps.UserCommands,
		queries:  deps.UserQueries,
	}
}

// Ensure UsersController implements handlers.UsersController.
var _ handlers.UsersController = (*UsersController)(nil)

func (c *UsersController) CreateUser(
	builder handlers.HandlerBuilder[*models.CreateUserParams, *models.CreateUserResponse],
) http.Handler {
	return builder.HandleWith(
		func(ctx context.Context, params *models.CreateUserParams) (*models.CreateUserResponse, error) {
			appReq := app.CreateUserRequest{
				Name:  params.Payload.Name,
				Email: params.Payload.Email,
			}
			res, err := c.commands.CreateUser(ctx, appReq)
			if err != nil {
				return nil, err
			}
			return &models.CreateUserResponse{UserID: res.UserID}, nil
		},
	)
}

func (c *UsersController) DeleteUser(builder handlers.NoResponseHandlerBuilder[*models.DeleteUserParams]) http.Handler {
	return builder.HandleWith(func(ctx context.Context, params *models.DeleteUserParams) error {
		return c.commands.DeleteUser(ctx, params.UserID)
	})
}

func (c *UsersController) GetUserByID(
	builder handlers.HandlerBuilder[*models.GetUserByIDParams, *models.UserResponse],
) http.Handler {
	return builder.HandleWith(
		func(ctx context.Context, params *models.GetUserByIDParams) (*models.UserResponse, error) {
			user, err := c.queries.GetUserByID(ctx, params.UserID)
			if err != nil {
				return nil, err
			}
			return &models.UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			}, nil
		},
	)
}

func (c *UsersController) ListUsers(builder handlers.NoParamsHandlerBuilder[*models.ListUsersResponse]) http.Handler {
	return builder.HandleWith(func(ctx context.Context) (*models.ListUsersResponse, error) {
		users, err := c.queries.ListUsers(ctx)
		if err != nil {
			return nil, err
		}
		userResponses := make([]*models.UserResponse, len(users))
		for i, user := range users {
			userResponses[i] = &models.UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			}
		}
		return &models.ListUsersResponse{Users: userResponses}, nil
	})
}

func (c *UsersController) UpdateUser(builder handlers.NoResponseHandlerBuilder[*models.UpdateUserParams]) http.Handler {
	return builder.HandleWith(
		func(ctx context.Context, params *models.UpdateUserParams) error {
			appReq := app.UpdateUserRequest{
				UserID: params.UserID,
				Name:   params.Payload.Name,
				Email:  params.Payload.Email,
			}
			return c.commands.UpdateUser(ctx, appReq)
		},
	)
}
