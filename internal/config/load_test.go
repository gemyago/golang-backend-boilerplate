package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("should load local config with default opts", func(t *testing.T) {
		cfg, err := Load(NewLoadOpts())
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.Equal(t, "DEBUG", cfg.GetString("defaultLogLevel"))
	})
	t.Run("should fail if no default config is found", func(t *testing.T) {
		opts := NewLoadOpts()
		opts.defaultConfigFileName = "not-existing.json"
		cfg, err := Load(opts)
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Nil(t, cfg)
	})
	t.Run("should load env specific config", func(t *testing.T) {
		cfg, err := Load(NewLoadOpts().WithEnv("test"))
		require.NoError(t, err)
		require.NotNil(t, cfg)

		require.Equal(t, "DEBUG", cfg.GetString("defaultLogLevel"))
	})
	t.Run("should return error if config is not found", func(t *testing.T) {
		cfg, err := Load(NewLoadOpts().WithEnv("not-existing"))
		require.ErrorIs(t, err, os.ErrNotExist)
		require.Nil(t, cfg)
	})
}
