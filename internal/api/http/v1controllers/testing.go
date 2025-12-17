//go:build !release

package v1controllers

import (
	"github.com/gemyago/golang-backend-boilerplate/internal/api/http/v1routes/models"
	"github.com/jaswdr/faker/v2"
)

const randomPrefixLength = 5

func newRandomCreateUserRequest(fake faker.Faker) models.CreateUserRequest {
	randomPrefix := fake.RandomStringWithLength(randomPrefixLength)
	return models.CreateUserRequest{
		Name:  "(" + randomPrefix + ") " + fake.Person().Name(),
		Email: randomPrefix + "." + fake.Internet().Email(),
	}
}

func newRandomUpdateUserRequest(fake faker.Faker) models.UpdateUserRequest {
	randomPrefix := fake.RandomStringWithLength(randomPrefixLength)
	return models.UpdateUserRequest{
		Name:  "(" + randomPrefix + ") " + fake.Person().Name(),
		Email: randomPrefix + "." + fake.Internet().Email(),
	}
}
