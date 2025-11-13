//go:build !release

package services

import (
	"time"

	"github.com/gemyago/golang-backend-boilerplate/internal/diag"
	"github.com/jaswdr/faker"
)

type MockNow struct {
	value time.Time
}

var _ TimeProvider = &MockNow{}

// NewMockNow constructor for MockNow.
func NewMockNow() *MockNow {
	fake := faker.New()
	return &MockNow{
		value: time.UnixMilli(fake.Time().Unix(time.Now())),
	}
}

func (m *MockNow) SetValue(t time.Time) {
	m.value = t
}

func (m *MockNow) Now() time.Time {
	return m.value
}

func MockNowValue(p TimeProvider) time.Time {
	mp, ok := p.(*MockNow)
	if !ok {
		panic("provided TimeProvider is not a MockNow")
	}
	return mp.value
}

const defaultTestShutdownTimeout = 30 * time.Second

func NewTestShutdownHooks() *ShutdownHooks {
	return NewShutdownHooks(ShutdownHooksRegistryDeps{
		RootLogger:              diag.RootTestLogger(),
		GracefulShutdownTimeout: defaultTestShutdownTimeout,
	})
}
