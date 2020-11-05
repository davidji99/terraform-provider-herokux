package platform

import "time"

// LogDrain represents a log drain in Heroku.
type LogDrain struct {
	Addon     *string    `json:"addon,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ID        *string    `json:"id,omitempty"`
	Token     *string    `json:"token,omitempty"`
	UpdatedAt *string    `json:"updated_at,omitempty"`
	URL       *string    `json:"url,omitempty"`
}
