package v1controllers

import (
	"github.com/stretchr/testify/mock"
)

// MockPetsCommands is a mock implementation of *app.PetsCommands.
type MockPetsCommands struct {
	mock.Mock
}

// MockPetsQueries is a mock implementation of *app.PetsQueries.
type MockPetsQueries struct {
	mock.Mock
}
