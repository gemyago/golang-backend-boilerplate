package v1controllers

import (
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
)

type UsersMapper struct{}

func (m *UsersMapper) MapUserToResponse(user *app.User) *models.UserResponse {
	return &models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
