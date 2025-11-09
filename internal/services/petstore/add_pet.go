package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
)

// AddPetParams contains parameters for adding a pet.
type AddPetParams struct {
	// Request represents the request body for adding a pet.
	Request *Pet
}

// AddPet adds a new pet to the store.
func (c *Client) AddPet(ctx context.Context, tokenProvider TokenProvider, params AddPetParams) (*Pet, error) {
	token, err := tokenProvider.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	ctxWithAuth := middleware.WithAuthTokenV2(ctx, token)

	// Make API call
	var response Pet
	err = http.SendRequest(ctxWithAuth, c.httpClient, http.SendRequestParams[Pet, Pet]{
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
