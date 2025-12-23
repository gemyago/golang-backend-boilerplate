package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/infrastructure/httpclient"
)

// UpdatePetParams contains parameters for updating a pet.
type UpdatePetParams struct {
	Request *Pet
}

func (c *Client) UpdatePet(ctx context.Context, params UpdatePetParams) (*Pet, error) {
	var response Pet
	err := httpclient.SendRequest(ctx, c.httpClient, httpclient.SendRequestParams[Pet, Pet]{
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
