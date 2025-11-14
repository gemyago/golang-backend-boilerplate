//go:build !release

package app

import (
	"github.com/jaswdr/faker"
)

func NewRandomCreateUserRequest(fake faker.Faker) *CreateUserRequest {
	randomPrefix := "(" + fake.RandomStringWithLength(5) + ") "
	return &CreateUserRequest{
		Name:  randomPrefix + fake.Person().Name(),
		Email: randomPrefix + fake.Internet().Email(),
	}
}
