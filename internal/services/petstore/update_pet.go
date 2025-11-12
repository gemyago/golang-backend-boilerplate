package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
	"golang.org/x/oauth2"
)

// UpdatePetParams contains parameters for updating a pet.
type UpdatePetParams struct {
	// Request represents the request body for updating a pet.
	Request *Pet
}

// UpdatePet updates an existing pet in the store.
func (c *Client) UpdatePet(
	ctx context.Context,
	tokenProvider oauth2.TokenSource,
	params UpdatePetParams,
) (*Pet, error) {
	token, err := tokenProvider.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	ctxWithAuth := middleware.WithAuthTokenV2(ctx, token)

	// Make API call
	var response Pet
	err = http.SendRequest(ctxWithAuth, c.httpClient, http.SendRequestParams[Pet, Pet]{
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
