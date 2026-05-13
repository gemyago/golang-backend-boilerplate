package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// HTTP status code threshold for errors.
	httpStatusErrorMin = 400
)

// SendRequestParams represents parameters for the SendRequest function.
// Use interface{} if no body and or target are needed.
type SendRequestParams[TBody any, TTarget any] struct {
	// HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string

	// Full URL for the request
	URL string

	// Request body to be JSON marshaled (can be nil for GET requests)
	Body *TBody

	// Target to unmarshal response into (can be nil if response not needed)
	Target *TTarget
}

// SendRequest performs an HTTP request with generic body and target types.
// This is a shared function that can be used by all API clients for consistent
// request handling, error processing, and response unmarshaling.
func SendRequest[TBody any, TTarget any](
	ctx context.Context,
	client *http.Client,
	params SendRequestParams[TBody, TTarget],
) error {
	var reqBody bytes.Buffer
	if params.Body != nil {
		if err := json.NewEncoder(&reqBody).Encode(params.Body); err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, params.Method, params.URL, &reqBody)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if params.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	//nolint:gosec // Request destination is supplied by the caller's configured client flow.
	resp, err := client.Do(req)
	if err != nil {
		return httpTransportErr(req, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= httpStatusErrorMin {
		return httpStatusErr(req, resp)
	}

	// We assume all responses are JSON. Adjust to handle different content types if needed.
	if params.Target != nil {
		if err = json.NewDecoder(resp.Body).Decode(params.Target); err != nil {
			return fmt.Errorf("failed to unmarshal response into target: %w", err)
		}
	}

	return nil
}

func httpTransportErr(req *http.Request, err error) *RequestError {
	return &RequestError{
		StatusCode: 0, // No status code for transport errors
		Method:     req.Method,
		URL:        req.URL.String(),
		Message:    "http transport error",
		Err:        err,
	}
}

func httpStatusErr(req *http.Request, resp *http.Response) *RequestError {
	message := fmt.Sprintf("http request error (%s)", resp.Status)

	// Response body may contain additional error details
	var bodyBytes []byte
	if resp.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			message += fmt.Sprintf("failed to read error response body: %v", err)
		}

		// Caller is closing the body
	}

	httpErr := &RequestError{
		StatusCode: resp.StatusCode,
		Method:     req.Method,
		URL:        req.URL.String(),
		Message:    message,
		Err:        nil, // No underlying error for HTTP status errors
		Body:       bodyBytes,
	}

	// We don't log here, it's caller layer responsibility

	return httpErr
}
