package main

import (
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	fake := faker.New()
	t.Run("http", func(t *testing.T) {
		t.Run("should initialize app", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SetArgs([]string{"http", "-e", "test", "--noop", "--logs-file", "../../test.log"})
			require.NoError(t, rootCmd.Execute())
		})
		t.Run("should fail if bad log level", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs(
				[]string{"http", "-e", "test", "--noop", "-l", fake.Lorem().Word(), "--logs-file", "../../test.log"},
			)
			assert.Error(t, rootCmd.Execute())
		})
		t.Run("should fail if unexpected env", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs([]string{"http", "--noop", "-e", fake.Lorem().Word(), "--logs-file", "../../test.log"})
			gotErr := rootCmd.Execute()
			assert.ErrorContains(t, gotErr, "failed to read config")
		})
	})
	t.Run("stdio", func(t *testing.T) {
		t.Run("should initialize app", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SetArgs([]string{"stdio", "-e", "test", "--noop", "--logs-file", "../../test.log"})
			require.NoError(t, rootCmd.Execute())
		})
		t.Run("should fail if bad log level", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs(
				[]string{"stdio", "-e", "test", "--noop", "-l", fake.Lorem().Word(), "--logs-file", "../../test.log"},
			)
			assert.Error(t, rootCmd.Execute())
		})
		t.Run("should fail if unexpected env", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs([]string{"stdio", "--noop", "-e", fake.Lorem().Word(), "--logs-file", "../../test.log"})
			gotErr := rootCmd.Execute()
			assert.ErrorContains(t, gotErr, "failed to read config")
		})
	})
}
