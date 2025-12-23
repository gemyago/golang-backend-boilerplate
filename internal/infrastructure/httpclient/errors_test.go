package httpclient

import (
	"errors"
	"testing"

	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/require"
)

func TestRequestError(t *testing.T) {
	f := faker.New()

	t.Run("Error_without_underlying_error", func(t *testing.T) {
		message := f.Lorem().Sentence(5)
		reqErr := &RequestError{
			Message: message,
		}
		require.Equal(t, message, reqErr.Error())
	})

	t.Run("Error_with_underlying_error_no_body", func(t *testing.T) {
		message := f.Lorem().Sentence(5)
		underlyingErr := errors.New(f.Lorem().Word())
		reqErr := &RequestError{
			Message: message,
			Err:     underlyingErr,
		}
		expected := message + ": " + underlyingErr.Error()
		require.Equal(t, expected, reqErr.Error())
	})

	t.Run("Error_with_underlying_error_and_body", func(t *testing.T) {
		message := f.Lorem().Sentence(5)
		underlyingErr := errors.New(f.Lorem().Word())
		body := []byte(f.Lorem().Paragraph(2))
		reqErr := &RequestError{
			Message: message,
			Err:     underlyingErr,
			Body:    body,
		}
		expected := message + "; response body: " + string(body) + ": " + underlyingErr.Error()
		require.Equal(t, expected, reqErr.Error())
	})

	t.Run("Unwrap_returns_underlying_error", func(t *testing.T) {
		underlyingErr := errors.New(f.Lorem().Word())
		reqErr := &RequestError{
			Err: underlyingErr,
		}
		require.Equal(t, underlyingErr, reqErr.Unwrap())
	})

	t.Run("Unwrap_returns_nil_when_no_underlying_error", func(t *testing.T) {
		reqErr := &RequestError{}
		require.NoError(t, reqErr.Unwrap())
	})
}
