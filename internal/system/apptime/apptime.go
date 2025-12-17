package apptime

import "time"

// Provider is a standard way to get current time across the system.
type Provider interface {
	Now() time.Time
}

// SystemProvider returns time based on system clock.
type SystemProvider struct{}

// NewSystemProvider returns an instance of SystemProvider.
func NewSystemProvider() *SystemProvider {
	return &SystemProvider{}
}

func (p *SystemProvider) Now() time.Time {
	return time.Now() // Implementation is based on system clock
}
