package petstore

import (
	"context"
	"fmt"

	"github.com/gemyago/golang-backend-boilerplate/internal/services/http"
	"github.com/gemyago/golang-backend-boilerplate/internal/services/http/middleware"
	"golang.org/x/oauth2"
)

// GetPetByIDParams contains parameters for getting a pet by ID.
type GetPetByIDParams struct {
	PetID int64
}

// GetPetByID retrieves a pet by its ID from the store.
func (c *Client) GetPetByID(
	ctx context.Context,
	tokenProvider oauth2.TokenSource,
	params GetPetByIDParams,
) (*Pet, error) {
	token, err := tokenProvider.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	ctxWithAuth := middleware.WithAuthTokenV2(ctx, token)

	var response Pet
	path := fmt.Sprintf("/pet/%d", params.PetID)
	err = http.SendRequest(ctxWithAuth, c.httpClient, http.SendRequestParams[interface{}, Pet]{
		Method: "GET",
		URL:    c.baseURL + path,
		Target: &response,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by ID: %w", err)
	}

	return &response, nil
}
