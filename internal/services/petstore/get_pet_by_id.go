package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http"
)

// GetPetByIDParams contains parameters for getting a pet by ID.
type GetPetByIDParams struct {
	PetID int64
}

// GetPetByID retrieves a pet by its ID from the store.
func (c *Client) GetPetByID(
	ctx context.Context,
	params GetPetByIDParams,
) (*Pet, error) {
	var response Pet
	path := fmt.Sprintf("/pet/%d", params.PetID)
	err := http.SendRequest(ctx, c.httpClient, http.SendRequestParams[interface{}, Pet]{
		Method: "GET",
		URL:    c.baseURL + path,
		Target: &response,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by ID: %w", err)
	}

	return &response, nil
}
