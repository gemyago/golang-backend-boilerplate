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

func TestClient_GetPetByID(t *testing.T) {
	fake := faker.New()

	t.Run("success with all parameters and fields", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := rand.Int64N(10) + 1 // 1 to 10 as per OpenAPI
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
		}

		// Prepare expected response with randomized data
		responseName := "pet-" + fake.Person().Name()
		responsePhotoUrls := []string{fake.Internet().URL(), fake.Internet().URL()}
		responseStatus := "available"

		expectedPet := Pet{
			ID:        petID,
			Name:      responseName,
			PhotoUrls: responsePhotoUrls,
			Status:    responseStatus,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request details
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, fmt.Sprintf("/pet/%d", petID), r.URL.Path)

			// Important to check token
			expectedAuth := mockTokenProvider.TokenType + " " + mockTokenProvider.TokenValue
			assert.Equal(t, expectedAuth, r.Header.Get("Authorization"))

			// Return complete successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf(`{
                "id": %d,
                "name": "%s",
                "photoUrls": ["%s", "%s"],
                "status": "%s"
            }`, petID, responseName, responsePhotoUrls[0], responsePhotoUrls[1], responseStatus)
			fmt.Fprint(w, response)
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
		assert.Equal(t, expectedPet, *pet)
	})

	t.Run("success with required parameters only", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := rand.Int64N(10) + 1
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
		}

		// Prepare expected minimal response with randomized data
		responseName := "minimal-pet-" + fake.Person().Name()
		responsePhotoUrls := []string{fake.Internet().URL()}

		expectedPet := Pet{
			ID:        petID,
			Name:      responseName,
			PhotoUrls: responsePhotoUrls,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Return minimal successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf(`{
                "id": %d,
                "name": "%s",
                "photoUrls": ["%s"]
            }`, petID, responseName, responsePhotoUrls[0])
			fmt.Fprint(w, response)
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
		assert.Equal(t, expectedPet, *pet)
	})

	t.Run("handles API error", func(t *testing.T) {
		// Arrange
		petID := rand.Int64N(10) + 1
		mockTokenProvider := &MockTokenProvider{
			TokenType:  fake.Lorem().Word(),
			TokenValue: fake.UUID().V4(),
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
			Err: errors.New(fake.Lorem().Sentence(10)),
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
