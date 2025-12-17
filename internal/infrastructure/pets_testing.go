//go:build !release

package infrastructure

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/jaswdr/faker/v2"
)

type RandomUserPetOpt func(*app.UserPet)

func WithUserPetUserID(userID string) RandomUserPetOpt {
	return func(up *app.UserPet) {
		up.UserID = userID
	}
}

func WithUserPetPetID(petID int64) RandomUserPetOpt {
	return func(up *app.UserPet) {
		up.PetID = petID
	}
}

func WithUserPetCreatedAt(createdAt time.Time) RandomUserPetOpt {
	return func(up *app.UserPet) {
		up.CreatedAt = createdAt
	}
}

func NewRandomUserPet(fake faker.Faker, opts ...RandomUserPetOpt) *app.UserPet {
	userPet := &app.UserPet{
		UserID:    fake.RandomStringWithLength(10),
		PetID:     fake.Int64Between(1, 10000),
		CreatedAt: fake.Time().Time(time.Now().Add(-time.Hour * pastYearHours)), // Random time in the past year
	}

	for _, opt := range opts {
		opt(userPet)
	}

	return userPet
}
