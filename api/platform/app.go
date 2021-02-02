package platform

// App represents a Heroku app.
type App struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}
