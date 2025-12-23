package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
)

// AddPetParams contains parameters for adding a pet.
type AddPetParams struct {
	// Request represents the request body for adding a pet.
	Request *Pet
}

// AddPet adds a new pet to the store.
func (c *Client) AddPet(ctx context.Context, params AddPetParams) (*Pet, error) {
	var response Pet
	err := httpclient.SendRequest(ctx, c.httpClient, httpclient.SendRequestParams[Pet, Pet]{
		Method: "POST",
		URL:    c.baseURL + "/pet",
		Body:   params.Request,
		Target: &response,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add pet: %w", err)
	}

	return &response, nil
}
