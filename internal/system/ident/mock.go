//go:build !release

package ident

import (
	"fmt"

	uuid "github.com/gofrs/uuid/v5"
)

type MockGenerator struct {
	upstream      Generator
	lastGenerated UUID
	nextGenerated UUID
}

var _ Generator = &MockGenerator{}

// NewMockGenerator creates a new MockGenerator with default upstream.
func NewMockGenerator() *MockGenerator {
	upstream := NewDefaultGenerator()
	return &MockGenerator{
		upstream:      upstream,
		nextGenerated: upstream.MustNewV7(),
	}
}

func (m *MockGenerator) MustNewV7() UUID {
	id, err := m.NewV7()
	if err != nil {
		panic(fmt.Errorf("failed to generate UUIDv7: %w", err))
	}
	return id
}

func (m *MockGenerator) NewV7() (UUID, error) {
	id, err := m.upstream.NewV7()
	if err != nil {
		return uuid.Nil, err
	}
	result := m.nextGenerated
	m.lastGenerated = result
	m.nextGenerated = id
	return result, nil
}

// MockGeneratorLastGenerated returns the last generated UUID from a MockGenerator.
func MockGeneratorLastGenerated(g Generator) UUID {
	mg, ok := g.(*MockGenerator)
	if !ok {
		panic("provided Generator is not a MockGenerator")
	}
	return mg.lastGenerated
}

// MockGeneratorNextGenerated returns the value that will be generated next.
func MockGeneratorNextGenerated(g Generator) UUID {
	mg, ok := g.(*MockGenerator)
	if !ok {
		panic("provided Generator is not a MockGenerator")
	}
	return mg.nextGenerated
}
