package http

import (
	"fmt"
)

// RequestError represents an HTTP-related error with context.
type RequestError struct {
	StatusCode int
	Method     string
	URL        string
	Message    string
	Err        error
	Body       []byte
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	if e.Err != nil {
		if e.Body != nil {
			return fmt.Sprintf("%s; response body: %s: %v", e.Message, string(e.Body), e.Err)
		}

		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements error unwrapping for error chain support.
func (e *RequestError) Unwrap() error {
	return e.Err
}
