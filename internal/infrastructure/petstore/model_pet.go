package petstore

// PetStatus represents the status of a pet.
type PetStatus string

const (
	PetStatusAvailable PetStatus = "available"
	PetStatusPending   PetStatus = "pending"
	PetStatusSold      PetStatus = "sold"
)

// Pet represents a pet.

type Pet struct {
	ID                 int64       `json:"id"`
	Category           *Category   `json:"category"`
	Name               string      `json:"name"`
	PhotoUrls          []string    `json:"photoUrls"`
	Tags               []*Tag      `json:"tags"`
	Status             PetStatus   `json:"status"`
	AvailableInstances int32       `json:"availableInstances"`
	PetDetailsID       int64       `json:"petDetailsId"`
	PetDetails         *PetDetails `json:"petDetails"`
}
