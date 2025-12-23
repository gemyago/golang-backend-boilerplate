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

type UserCommands interface {
	CreateUser(ctx context.Context, req app.CreateUserRequest) (*app.CreateUserResponse, error)
	UpdateUser(ctx context.Context, req app.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID string) error
}

// Ensure interface is compatible.
var _ UserCommands = (*app.UserCommands)(nil)

type UserQueries interface {
	GetUserByID(ctx context.Context, userID string) (*app.User, error)
	ListUsers(ctx context.Context) ([]*app.User, error)
}

// Ensure interface is compatible.
var _ UserQueries = (*app.UserQueries)(nil)

type UsersController struct {
	commands UserCommands
	queries  UserQueries
	mapper   *UsersMapper
}

type UsersControllerDeps struct {
	dig.In

	UserCommands
	UserQueries
	*UsersMapper

	RootLogger *slog.Logger
}

func newUsersController(deps UsersControllerDeps) *UsersController {
	return &UsersController{
		commands: deps.UserCommands,
		queries:  deps.UserQueries,
		mapper:   deps.UsersMapper,
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
			return c.mapper.MapUserToResponse(user), nil
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
			userResponses[i] = c.mapper.MapUserToResponse(user)
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
