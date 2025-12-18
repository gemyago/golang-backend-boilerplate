package ident

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenerator(t *testing.T) {
	t.Run("NewV7", func(t *testing.T) {
		gen := NewDefaultGenerator()
		id, err := gen.NewV7()
		require.NoError(t, err)
		assert.NotEqual(t, UUID{}, id)
		assert.EqualValues(t, 7, id.Version())
	})

	t.Run("MustNewV7", func(t *testing.T) {
		gen := NewDefaultGenerator()
		id := gen.MustNewV7()
		assert.NotEqual(t, UUID{}, id)
		assert.EqualValues(t, 7, id.Version())
	})

	t.Run("Uniqueness", func(t *testing.T) {
		gen := NewDefaultGenerator()
		id1 := gen.MustNewV7()
		id2 := gen.MustNewV7()
		assert.NotEqual(t, id1, id2)
	})
}
