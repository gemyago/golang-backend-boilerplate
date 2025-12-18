package v1controllers

import (
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
)

func TestUsersMapper_MapUserToResponse(t *testing.T) {
	fake := faker.New()
	mapper := &UsersMapper{}

	t.Run("maps app.User to models.UserResponse correctly", func(t *testing.T) {
		user := app.User{
			ID:    fake.UUID().V4(),
			Name:  fake.Person().Name(),
			Email: fake.Internet().Email(),
		}

		result := mapper.MapUserToResponse(&user)

		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, user.Name, result.Name)
		assert.Equal(t, user.Email, result.Email)
	})
}

var _ *models.UserResponse
