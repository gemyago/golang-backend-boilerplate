package petstore

// PetDetails represents detailed information about a pet.
type PetDetails struct {
	ID       int64     `json:"id,omitempty"`
	Category *Category `json:"category,omitempty"`
	Tag      *Tag      `json:"tag,omitempty"`
}
