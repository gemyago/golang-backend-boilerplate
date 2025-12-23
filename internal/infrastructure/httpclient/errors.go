package httpclient

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
	msg := e.Message
	if e.Body != nil {
		msg += "; response body: " + string(e.Body)
	}

	if e.Err != nil {
		if e.Body != nil {
			return fmt.Sprintf("%s: %v", msg, e.Err)
		}

		return fmt.Sprintf("%s: %v", msg, e.Err)
	}
	return msg
}

// Unwrap implements error unwrapping for error chain support.
func (e *RequestError) Unwrap() error {
	return e.Err
}
