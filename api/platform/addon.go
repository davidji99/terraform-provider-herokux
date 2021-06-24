package platform

type Addon struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	App  *App    `json:"app,omitempty"`
}
