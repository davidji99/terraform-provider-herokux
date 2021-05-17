package platform

// Permission represents a permission.
type Permission struct {
	Name *string `json:"name,omitempty"`
}

// PermissionNotFoundError is returned when attempting to retrieve permissions that do not exist on a Heroku resource.
type PermissionNotFoundError struct {
	error
}
