package petstore

import (
	"context"
	"fmt"

	httpsvc "github.com/gemyago/golang-backend-boilerplate/internal/services/http"
)

// UpdatePetParams contains parameters for updating a pet.
type UpdatePetParams struct {
	Request *Pet
}

func (c *Client) UpdatePet(ctx context.Context, params UpdatePetParams) (*Pet, error) {
	var response Pet
	err := httpsvc.SendRequest(ctx, c.httpClient, httpsvc.SendRequestParams[Pet, Pet]{
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
