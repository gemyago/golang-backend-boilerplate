package infrastructure

import (
	"os"
	"path"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/system/lifecycle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupDatabase(t *testing.T) {
	t.Run("should setup the database with all tables", func(t *testing.T) {
		db, err := newDBProvider(t.Context())(DatabaseDeps{
			DSN:           ":memory:",
			ShutdownHooks: lifecycle.NewTestShutdownHooks(),
		})
		require.NoError(t, err)

		// Collect all table names from the database
		var allTableNames []string
		rows, err := db.instance.QueryContext(t.Context(),
			"SELECT name FROM sqlite_master WHERE type='table'")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var tableName string
			require.NoError(t, rows.Scan(&tableName))
			allTableNames = append(allTableNames, tableName)
		}

		require.NoError(t, rows.Err())

		// Verify that required tables exist
		require.Contains(t, allTableNames, "users")
	})

	t.Run("should return error if db is readonly (schema init will fail)", func(t *testing.T) {
		dbFilePath := path.Join(t.TempDir(), "invalid-db-name")
		file, err := os.Create(dbFilePath)
		require.NoError(t, err)
		file.Close()

		require.NoError(t, os.Chmod(dbFilePath, 0400))

		_, err = newDBProvider(t.Context())(DatabaseDeps{
			DSN:           dbFilePath,
			ShutdownHooks: lifecycle.NewTestShutdownHooks(),
		})
		require.Error(t, err)
		assert.Regexp(t, `failed to initialize schema`, err.Error())
	})
}
