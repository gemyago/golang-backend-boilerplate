package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRequest(t *testing.T) {
	fake := faker.New()

	t.Run("GET request with response target", func(t *testing.T) {
		// Create test server that returns a JSON response
		userID := fake.UUID().V4()
		userName := fake.Person().Name()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/users/"+userID, r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"id":"%s","name":"%s"}`, userID, userName)
		}))
		defer server.Close()

		client := server.Client()
		ctx := t.Context()

		type ResponseData struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		var response ResponseData
		params := SendRequestParams[any, ResponseData]{
			Method: "GET",
			URL:    server.URL + "/users/" + userID,
			Body:   nil,
			Target: &response,
		}

		err := SendRequest(ctx, client, params)

		require.NoError(t, err)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, userName, response.Name)
	})

	t.Run("POST request with body and response", func(t *testing.T) {
		// Create test server that accepts POST and returns response
		userID := fake.UUID().V4()
		userName := fake.Person().Name()
		userEmail := fake.Internet().Email()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/users", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"id":"%s","name":"%s","email":"%s"}`, userID, userName, userEmail)
		}))
		defer server.Close()

		client := server.Client()
		ctx := t.Context()

		type RequestData struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type ResponseData struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		requestBody := RequestData{
			Name:  userName,
			Email: userEmail,
		}

		var response ResponseData
		params := SendRequestParams[RequestData, ResponseData]{
			Method: "POST",
			URL:    server.URL + "/users",
			Body:   &requestBody,
			Target: &response,
		}

		err := SendRequest(ctx, client, params)

		require.NoError(t, err)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, userName, response.Name)
		assert.Equal(t, userEmail, response.Email)
	})

	t.Run("DELETE request with no body or response", func(t *testing.T) {
		// Create test server that handles DELETE
		userID := fake.UUID().V4()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "DELETE", r.Method)
			assert.Equal(t, "/users/"+userID, r.URL.Path)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := server.Client()
		ctx := t.Context()

		params := SendRequestParams[any, any]{
			Method: "DELETE",
			URL:    server.URL + "/users/" + userID,
			Body:   nil,
			Target: nil,
		}

		err := SendRequest(ctx, client, params)

		require.NoError(t, err)
	})

	t.Run("handles HTTP error responses", func(t *testing.T) {
		// Create test server that returns error
		nonExistentPath := "/" + fake.Lorem().Word()
		wantErrorBody := `{"error":"` + fake.Lorem().Word() + `"}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, wantErrorBody)
		}))
		defer server.Close()

		client := server.Client()
		client.Transport = NewClientFactory(ClientFactoryDeps{
			RootLogger: telemetry.RootTestLogger(),
		}).CreateClient().Transport
		ctx := t.Context()

		params := SendRequestParams[any, any]{
			Method: "GET",
			URL:    server.URL + nonExistentPath,
			Body:   nil,
			Target: nil,
		}

		err := SendRequest(ctx, client, params)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")

		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Equal(t, http.StatusInternalServerError, reqErr.StatusCode)
		assert.Equal(t, "GET", reqErr.Method)
		assert.Equal(t, server.URL+nonExistentPath, reqErr.URL)
		assert.Contains(t, string(reqErr.Body), wantErrorBody)
	})

	t.Run("handles invalid URL", func(t *testing.T) {
		client := &http.Client{}
		ctx := t.Context()

		params := SendRequestParams[any, any]{
			Method: "GET",
			URL:    "not-a-valid-url",
			Body:   nil,
			Target: nil,
		}

		err := SendRequest(ctx, client, params)

		require.Error(t, err)

		var reqErr *RequestError
		require.ErrorAs(t, err, &reqErr)
		assert.Equal(t, "GET", reqErr.Method)
		assert.Equal(t, "not-a-valid-url", reqErr.URL)
	})

	t.Run("handles request body marshaling error", func(t *testing.T) {
		client := &http.Client{}
		ctx := t.Context()

		// Use a type that can't be marshaled to JSON
		type InvalidBody struct {
			Channel chan int `json:"channel"` // channels can't be marshaled
		}

		invalidBody := InvalidBody{
			Channel: make(chan int),
		}

		params := SendRequestParams[InvalidBody, any]{
			Method: "POST",
			URL:    "http://example.com",
			Body:   &invalidBody,
			Target: nil,
		}

		err := SendRequest(ctx, client, params)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal request body")
	})

	t.Run("handles response unmarshaling error", func(t *testing.T) {
		// Create test server that returns invalid JSON
		invalidPath := "/" + fake.Lorem().Word()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"invalid": json}`) // Invalid JSON
		}))
		defer server.Close()

		client := server.Client()
		ctx := t.Context()

		type ResponseData struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		var response ResponseData
		params := SendRequestParams[any, ResponseData]{
			Method: "GET",
			URL:    server.URL + invalidPath,
			Body:   nil,
			Target: &response,
		}

		err := SendRequest(ctx, client, params)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal response")
	})
}
