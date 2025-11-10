package petstore

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_AddPet(t *testing.T) {
	fake := faker.New()

	t.Run("success with all parameters and fields", func(t *testing.T) {
		// Arrange - Use randomized data
		petName := fake.Person().Name()
		photoUrls := []string{fake.Internet().URL(), fake.Internet().URL()}
		status := "available"
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
		}

		// Prepare expected response with randomized data
		responseID := rand.Int64N(1000) + 1
		responseName := "pet-" + fake.Person().Name()
		responsePhotoUrls := []string{fake.Internet().URL(), fake.Internet().URL()}
		responseStatus := "available"

		expectedPet := Pet{
			ID:        responseID,
			Name:      responseName,
			PhotoUrls: responsePhotoUrls,
			Status:    responseStatus,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request details
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/pet", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Important to check token
			expectedAuth := mockTokenProvider.TokenType + " " + mockTokenProvider.TokenValue
			assert.Equal(t, expectedAuth, r.Header.Get("Authorization"))

			// Return complete successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
                "id": %d,
                "name": "%s",
                "photoUrls": ["%s", "%s"],
                "status": "%s"
            }`, responseID, responseName, responsePhotoUrls[0], responsePhotoUrls[1], responseStatus)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			Name:      petName,
			PhotoUrls: photoUrls,
			Status:    status,
		}

		// Act
		pet, err := client.AddPet(t.Context(), mockTokenProvider, AddPetParams{
			Request: req,
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedPet, *pet)
	})

	t.Run("success with required parameters only", func(t *testing.T) {
		// Arrange - Use randomized data
		petName := fake.Person().Name()
		photoUrls := []string{fake.Internet().URL()}
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
		}

		// Prepare expected minimal response with randomized data
		responseID := rand.Int64N(1000) + 1
		responseName := "minimal-pet-" + fake.Person().Name()
		responsePhotoUrls := []string{fake.Internet().URL()}

		expectedPet := Pet{
			ID:        responseID,
			Name:      responseName,
			PhotoUrls: responsePhotoUrls,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Return minimal successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
                "id": %d,
                "name": "%s",
                "photoUrls": ["%s"]
            }`, responseID, responseName, responsePhotoUrls[0])
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			Name:      petName,
			PhotoUrls: photoUrls,
		}

		// Act
		pet, err := client.AddPet(t.Context(), mockTokenProvider, AddPetParams{
			Request: req,
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedPet, *pet)
	})

	t.Run("handles API error", func(t *testing.T) {
		// Arrange
		petName := fake.Person().Name()
		photoUrls := []string{fake.Internet().URL()}
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			Name:      petName,
			PhotoUrls: photoUrls,
		}

		// Act
		result, err := client.AddPet(t.Context(), mockTokenProvider, AddPetParams{
			Request: req,
		})

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to add pet")
	})

	t.Run("handles token provider error", func(t *testing.T) {
		// Arrange
		petName := fake.Person().Name()
		photoUrls := []string{fake.Internet().URL()}
		mockTokenProvider := &MockTokenProvider{
			Err: errors.New(fake.Lorem().Sentence(10)),
		}

		deps := makeMockDeps(t, "http://example.com")
		client := NewClient(deps)

		req := &Pet{
			Name:      petName,
			PhotoUrls: photoUrls,
		}

		// Act
		result, err := client.AddPet(t.Context(), mockTokenProvider, AddPetParams{
			Request: req,
		})

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		expectedError := fmt.Errorf("failed to get token: %w", mockTokenProvider.Err)
		assert.Equal(t, expectedError.Error(), err.Error())
	})
}
