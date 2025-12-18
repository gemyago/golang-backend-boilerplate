package apptime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSystemProvider_Now(t *testing.T) {
	tp := NewSystemProvider()
	now := time.Now()
	got := tp.Now()
	assert.WithinDuration(t, now, got, time.Second)
}
