package petstore

// Pet represents a pet.
type Pet struct {
	ID                 int64       `json:"id,omitempty"`
	Category           *Category   `json:"category,omitempty"`
	Name               string      `json:"name"`
	PhotoUrls          []string    `json:"photoUrls"`
	Tags               []Tag       `json:"tags,omitempty"`
	Status             string      `json:"status,omitempty"`
	AvailableInstances *int32      `json:"availableInstances,omitempty"`
	PetDetailsID       int64       `json:"petDetailsId,omitempty"`
	PetDetails         *PetDetails `json:"petDetails,omitempty"`
}
