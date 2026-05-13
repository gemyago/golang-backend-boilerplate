package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
)

// GetPetByIDParams contains parameters for getting a pet by ID.
type GetPetByIDParams struct {
	PetID string
}

// GetPetByID retrieves a pet by its ID from the store.
func (c *Client) GetPetByID(ctx context.Context, params GetPetByIDParams) (*Pet, error) {
	var response Pet
	path := fmt.Sprintf("/pet/%s", params.PetID)
	err := httpclient.SendRequest(ctx, c.httpClient, httpclient.SendRequestParams[any, Pet]{
		Method: "GET",
		URL:    c.baseURL + path,
		Target: &response,
	})
	if err != nil {
		return nil, fmt.Errorf("get pet by id failed: %w", err)
	}

	return &response, nil
}
