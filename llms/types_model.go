package llms

type Model struct {
	ID      string `json:"id"`
	Created int64  `json:"created,omitempty"`
	Object  string `json:"object,omitempty"`
	Ownedby string `json:"ownedby,omitempty"`

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
