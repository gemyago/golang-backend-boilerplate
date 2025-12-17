//go:build !release

package infrastructure

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/app"
	"github.com/gemyago/golang-backend-boilerplate/internal/system/ident"
	"github.com/jaswdr/faker/v2"
)

const pastYearHours = 24 * 365

type RandomUserOpt func(*app.User)

func WithUserID(id string) RandomUserOpt {
	return func(u *app.User) {
		u.ID = id
	}
}

func WithUserName(name string) RandomUserOpt {
	return func(u *app.User) {
		u.Name = name
	}
}

func WithUserEmail(email string) RandomUserOpt {
	return func(u *app.User) {
		u.Email = email
	}
}

func WithUserTimestamps(createdAt, updatedAt time.Time) RandomUserOpt {
	return func(u *app.User) {
		u.CreatedAt = createdAt
		u.UpdatedAt = updatedAt
	}
}

func NewRandomUser(fake faker.Faker, opts ...RandomUserOpt) *app.User {
	randomPrefix := "(" + fake.RandomStringWithLength(5) + ") "
	user := &app.User{
		ID:        ident.NewDefaultGenerator().MustNewV7().String(),
		Name:      randomPrefix + fake.Person().Name(),
		Email:     randomPrefix + fake.Internet().Email(),
		CreatedAt: fake.Time().Time(time.Now().Add(-time.Hour * pastYearHours)), // Random time in the past year
		UpdatedAt: fake.Time().Time(time.Now().Add(-time.Hour * pastYearHours)),
	}

	for _, opt := range opts {
		opt(user)
	}

	return user
}
