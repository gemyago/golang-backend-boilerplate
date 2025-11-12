package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http"
)

// UpdatePetParams contains parameters for updating a pet.
type UpdatePetParams struct {
	// Request represents the request body for updating a pet.
	Request *Pet
}

// UpdatePet updates an existing pet in the store.
func (c *Client) UpdatePet(
	ctx context.Context,
	params UpdatePetParams,
) (*Pet, error) {
	// Make API call
	var response Pet
	err := http.SendRequest(ctx, c.httpClient, http.SendRequestParams[Pet, Pet]{
		Method: "PUT",
		URL:    c.baseURL + "/pet",
		Body:   params.Request,
		Target: &response,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update pet: %w", err)
	}

	return &response, nil
}
