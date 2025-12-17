//go:build !release

package app

import (
	"github.com/jaswdr/faker/v2"
)

func NewRandomAddPetRequest(fake faker.Faker) *AddPetRequest {
	randomPrefix := fake.RandomStringWithLength(5)
	return &AddPetRequest{
		UserID:    fake.UUID().V4(),
		Name:      "(" + randomPrefix + ") " + fake.Person().Name() + "'s Pet",
		Status:    "available",
		PhotoUrls: []string{fake.Internet().URL()},
	}
}
