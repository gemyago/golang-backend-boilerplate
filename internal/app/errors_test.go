package app

import (
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomErrors(t *testing.T) {
	fake := faker.New()

	t.Run("NotFoundError", func(t *testing.T) {
		t.Run("should create error with resource and ID", func(t *testing.T) {
			resource := fake.Lorem().Word()
			id := fake.UUID().V4()
			err := NewErrNotFound(resource, id)

			expectedMsg := resource + " not found: " + id
			assert.Equal(t, expectedMsg, err.Error())
			assert.Equal(t, resource, err.Resource)
			assert.Equal(t, id, err.ID)
		})

		t.Run("should be detectable with errors.As", func(t *testing.T) {
			resource := fake.Lorem().Word()
			id := fake.UUID().V4()
			err := NewErrNotFound(resource, id)

			var notFoundErr *NotFoundError
			require.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, resource, notFoundErr.Resource)
			assert.Equal(t, id, notFoundErr.ID)
		})
	})

	t.Run("InvalidInputError", func(t *testing.T) {
		t.Run("should create error with field and reason", func(t *testing.T) {
			field := fake.Lorem().Word()
			reason := fake.Lorem().Sentence(3)
			err := NewErrInvalidInput(field, reason)

			expectedMsg := "invalid input for field '" + field + "': " + reason
			assert.Equal(t, expectedMsg, err.Error())
			assert.Equal(t, field, err.Field)
			assert.Equal(t, reason, err.Reason)
		})

		t.Run("should be detectable with errors.As", func(t *testing.T) {
			field := fake.Lorem().Word()
			reason := fake.Lorem().Sentence(3)
			err := NewErrInvalidInput(field, reason)

			var invalidInputErr *InvalidInputError
			require.ErrorAs(t, err, &invalidInputErr)
			assert.Equal(t, field, invalidInputErr.Field)
			assert.Equal(t, reason, invalidInputErr.Reason)
		})
	})

	t.Run("ConflictError", func(t *testing.T) {
		t.Run("should create error with resource and reason", func(t *testing.T) {
			resource := fake.Lorem().Word()
			reason := fake.Lorem().Sentence(3)
			err := NewErrConflict(resource, reason)

			expectedMsg := "conflict with " + resource + ": " + reason
			assert.Equal(t, expectedMsg, err.Error())
			assert.Equal(t, resource, err.Resource)
			assert.Equal(t, reason, err.Reason)
		})

		t.Run("should be detectable with errors.As", func(t *testing.T) {
			resource := fake.Lorem().Word()
			reason := fake.Lorem().Sentence(3)
			err := NewErrConflict(resource, reason)

			var conflictErr *ConflictError
			require.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, resource, conflictErr.Resource)
			assert.Equal(t, reason, conflictErr.Reason)
		})
	})
}
