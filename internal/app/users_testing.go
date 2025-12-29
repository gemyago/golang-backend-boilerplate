//go:build !release

package app

import (
	"time"

	"github.com/jaswdr/faker/v2"
)

type UserOption func(*User)

func WithUserID(id string) UserOption {
	return func(u *User) {
		u.ID = id
	}
}

func WithUserTimestamps(createdAt, updatedAt time.Time) UserOption {
	return func(u *User) {
		u.CreatedAt = createdAt.Truncate(time.Millisecond)
		u.UpdatedAt = updatedAt.Truncate(time.Millisecond)
	}
}

func WithUserEmail(email string) UserOption {
	return func(u *User) {
		u.Email = email
	}
}

func NewRandomUser(fake faker.Faker, opts ...UserOption) *User {
	user := &User{
		ID:        fake.RandomStringWithLength(10),
		Name:      fake.Person().Name(),
		Email:     fake.Internet().Email(),
		CreatedAt: fake.Time().Past(),
		UpdatedAt: fake.Time().Recent(),
	}
	for _, opt := range opts {
		opt(user)
	}
	return user
}

func NewRandomCreateUserRequest(fake faker.Faker) *CreateUserRequest {
	randomPrefix := fake.RandomStringWithLength(5)
	return &CreateUserRequest{
		Name:  "(" + randomPrefix + ") " + fake.Person().Name(),
		Email: randomPrefix + "." + fake.Internet().Email(),
	}
}

func NewRandomUpdateUserRequest(fake faker.Faker) *UpdateUserRequest {
	randomPrefix := fake.RandomStringWithLength(5)
	return &UpdateUserRequest{
		UserID: fake.RandomStringWithLength(10),
		Name:   "(" + randomPrefix + ") " + fake.Person().Name(),
		Email:  randomPrefix + "." + fake.Internet().Email(),
	}
}
