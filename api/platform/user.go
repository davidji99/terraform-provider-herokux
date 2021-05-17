package platform

// User represents a user in Heroku.
type User struct {
	ID    *string `json:"id"`
	Email *string `json:"email"`
}
