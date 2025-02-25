package llms

// Model represents a language model in the LLM system.
// It contains metadata and identification information for a specific model.
type Model struct {
	// ID is the unique identifier for the model
	ID      string `json:"id"`
	// Created is the timestamp when the model was created
	Created int64  `json:"created,omitempty"`
	// Object is the type of object (typically 'model')
	Object  string `json:"object,omitempty"`
	// Ownedby indicates the owner or organization that created the model
	Ownedby string `json:"ownedby,omitempty"`

	// Name is the display name of the model
	Name        string `json:"name,omitempty"`
	// Description provides additional information about the model
	Description string `json:"description,omitempty"`
}
