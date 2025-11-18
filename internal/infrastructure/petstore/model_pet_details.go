package petstore

// PetDetails represents details of a pet.
type PetDetails struct {
	ID       int64     `json:"id"`
	Category *Category `json:"category"`
	Tag      *Tag      `json:"tag"`
}
