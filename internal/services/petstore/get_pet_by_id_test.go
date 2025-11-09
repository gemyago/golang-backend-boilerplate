package petstore

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetPetByID(t *testing.T) {
	t.Run("success with all parameters and fields", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := rand.Int64N(10) + 1 // 1 to 10 as per OpenAPI
		mockTokenProvider := &MockTokenProvider{
			TokenType:  faker.Word(),
			TokenValue: faker.UUIDHyphenated(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, fmt.Sprintf("/pet/%d", petID), req.URL.Path)

			// Important to check token
			expectedAuth := mockTokenProvider.TokenType + " " + mockTokenProvider.TokenValue
			assert.Equal(t, expectedAuth, req.Header.Get("Authorization"))

			// Return complete successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
                "id": 123,
                "name": "test-pet",
                "photoUrls": ["url1", "url2"],
                "status": "available"
            }`)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		// Act
		pet, err := client.GetPetByID(t.Context(), mockTokenProvider, GetPetByIDParams{
			PetID: petID,
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(123), pet.ID)
		assert.Equal(t, "test-pet", pet.Name)
		assert.Equal(t, []string{"url1", "url2"}, pet.PhotoUrls)
		assert.Equal(t, "available", pet.Status)
	})

	t.Run("success with required parameters only", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := rand.Int64N(10) + 1
		mockTokenProvider := &MockTokenProvider{
			TokenType:  faker.Word(),
			TokenValue: faker.UUIDHyphenated(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Return minimal successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
                "id": 456,
                "name": "minimal-pet",
                "photoUrls": ["url"]
            }`)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		// Act
		pet, err := client.GetPetByID(t.Context(), mockTokenProvider, GetPetByIDParams{
			PetID: petID,
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, int64(456), pet.ID)
		assert.Equal(t, "minimal-pet", pet.Name)
		assert.Equal(t, []string{"url"}, pet.PhotoUrls)
	})

	t.Run("handles API error", func(t *testing.T) {
		// Arrange
		petID := rand.Int64N(10) + 1
		mockTokenProvider := &MockTokenProvider{
			TokenType:  faker.Word(),
			TokenValue: faker.UUIDHyphenated(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		// Act
		result, err := client.GetPetByID(t.Context(), mockTokenProvider, GetPetByIDParams{
			PetID: petID,
		})

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to get pet by ID")
	})

	t.Run("handles token provider error", func(t *testing.T) {
		// Arrange
		petID := rand.Int64N(10) + 1
		mockTokenProvider := &MockTokenProvider{
			Err: errors.New(faker.Sentence()),
		}

		deps := makeMockDeps(t, "http://example.com")
		client := NewClient(deps)

		// Act
		result, err := client.GetPetByID(t.Context(), mockTokenProvider, GetPetByIDParams{
			PetID: petID,
		})

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		expectedError := fmt.Errorf("failed to get token: %w", mockTokenProvider.Err)
		assert.Equal(t, expectedError.Error(), err.Error())
	})
}
