package petstore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
	"github.com/gemyago/golang-backend-boilerplate/internal/telemetry"
	"github.com/jaswdr/faker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_UpdatePet(t *testing.T) {
	makeMockDeps := func(t *testing.T, baseURL string) ClientDeps {
		// Always include test name in the logger for better debugging
		rootLogger := telemetry.RootTestLogger().With("test", t.Name())
		return ClientDeps{
			ClientFactory: httpclient.NewClientFactory(httpclient.ClientFactoryDeps{
				RootLogger: rootLogger,
			}),
			RootLogger: rootLogger,
			BaseURL:    baseURL,
		}
	}

	fake := faker.New()

	t.Run("success with all parameters and fields", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := fake.Int64()
		petName := fake.Person().Name()
		petPhotoUrls := []string{fake.Internet().URL(), fake.Internet().URL()}
		petCategory := &Category{ID: fake.Int64(), Name: fake.Lorem().Word()}
		petTags := []*Tag{{ID: fake.Int64(), Name: fake.Lorem().Word()}, {ID: fake.Int64(), Name: fake.Lorem().Word()}}
		petStatus := PetStatusAvailable
		petAvailableInstances := fake.Int32()
		petDetailsID := fake.Int64()
		petDetails := &PetDetails{ID: fake.Int64(), Category: petCategory, Tag: petTags[0]}

		expectedResponse := &Pet{
			ID:                 petID,
			Category:           petCategory,
			Name:               petName,
			PhotoUrls:          petPhotoUrls,
			Tags:               petTags,
			Status:             petStatus,
			AvailableInstances: petAvailableInstances,
			PetDetailsID:       petDetailsID,
			PetDetails:         petDetails,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request details
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/pet", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Return complete successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintf(w, `{
				"id": %d,
				"category": {"id": %d, "name": "%s"},
				"name": "%s",
				"photoUrls": ["%s", "%s"],
				"tags": [{"id": %d, "name": "%s"}, {"id": %d, "name": "%s"}],
				"status": "%s",
				"availableInstances": %d,
				"petDetailsId": %d,
				"petDetails": {"id": %d, "category": {"id": %d, "name": "%s"}, "tag": {"id": %d, "name": "%s"}}
			}`, petID, petCategory.ID, petCategory.Name, petName, petPhotoUrls[0], petPhotoUrls[1],
				petTags[0].ID, petTags[0].Name, petTags[1].ID, petTags[1].Name, petStatus,
				petAvailableInstances, petDetailsID, petDetails.ID, petCategory.ID, petCategory.Name,
				petTags[0].ID, petTags[0].Name)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			ID:                 petID,
			Category:           petCategory,
			Name:               petName,
			PhotoUrls:          petPhotoUrls,
			Tags:               petTags,
			Status:             petStatus,
			AvailableInstances: petAvailableInstances,
			PetDetailsID:       petDetailsID,
			PetDetails:         petDetails,
		}

		// Act
		pet, err := client.UpdatePet(t.Context(), UpdatePetParams{
			Request: req,
		})

		// Assert
		require.NoError(t, err)
		// Compare entire structs to avoid field-by-field assertions
		assert.Equal(t, expectedResponse, pet)
	})

	t.Run("success with required parameters only", func(t *testing.T) {
		// Arrange - Use randomized data
		petID := fake.Int64()
		petName := fake.Person().Name()
		petPhotoUrls := []string{fake.Internet().URL()}

		// Prepare expected minimal response with randomized data
		expectedName := fake.Person().Name()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Return minimal successful response with randomized data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"id": %d,
				"name": "%s",
				"photoUrls": ["%s"]
			}`, petID, expectedName, petPhotoUrls[0])
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			ID:        petID,
			Name:      petName, // Required fields
			PhotoUrls: petPhotoUrls,
		}

		// Act
		pet, err := client.UpdatePet(t.Context(), UpdatePetParams{
			Request: req,
		})

		// Assert
		require.NoError(t, err)
		// Check only the fields that should be present in minimal response
		assert.Equal(t, petID, pet.ID)
		assert.Equal(t, expectedName, pet.Name)
		assert.Equal(t, petPhotoUrls, pet.PhotoUrls)
	})

	t.Run("handles API error", func(t *testing.T) {
		// Arrange
		petID := fake.Int64()
		petName := fake.Person().Name()
		petPhotoUrls := []string{fake.Internet().URL()}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		deps := makeMockDeps(t, server.URL)
		client := NewClient(deps)

		req := &Pet{
			ID:        petID,
			Name:      petName,
			PhotoUrls: petPhotoUrls,
		}

		// Act
		result, err := client.UpdatePet(t.Context(), UpdatePetParams{
			Request: req,
		})

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorContains(t, err, "failed to update pet")
	})
}
