//go:build !release

package apptime

import (
	"time"

	"github.com/jaswdr/faker/v2"
)

type MockProvider struct {
	value time.Time
}

var _ Provider = &MockProvider{}

// NewMockProvider constructor for MockProvider.
func NewMockProvider() *MockProvider {
	fake := faker.New()
	return &MockProvider{
		value: time.UnixMilli(fake.Time().Unix(time.Now())),
	}
}

func (m *MockProvider) SetValue(t time.Time) {
	m.value = t
}

func (m *MockProvider) Now() time.Time {
	return m.value
}

func MockProviderValue(p Provider) time.Time {
	mp, ok := p.(*MockProvider)
	if !ok {
		panic("provided Provider is not a MockProvider")
	}
	return mp.value
}

func SetMockProviderValue(p Provider, val time.Time) {
	mp, ok := p.(*MockProvider)
	if !ok {
		panic("provided Provider is not a MockProvider")
	}
	mp.SetValue(val)
}
