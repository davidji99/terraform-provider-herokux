package graphql

import (
	"encoding/json"
	"strings"
)

// Request represents a Graphql request.
type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response represents a GraphQL response.
type Response struct {
	Data interface{} `json:"data"`
}

// Error represents a GraphQL error.
type Error struct {
	Message            string               `json:"message"`
	DownstreamResponse []DownstreamResponse `json:"downstreamResponse,omitempty"`
}

// DownstreamResponse represents an error's downstream response.
type DownstreamResponse struct {
	Locations []struct {
		Line   int `json:"line,omitempty"`
		Column int `json:"column,omitempty"`
	} `json:"locations,omitempty"`
	Path []string `json:"path,omitempty"`
}

// ErrorResponse represents the default GraphQL error response body.
type ErrorResponse struct {
	Errors     []Error `json:"errors,omitempty"`
	Extensions struct {
		Code             string `json:"code,omitempty"`
		ValidationErrors []struct {
			Name   string `json:"name,omitempty"`
			Reason string `json:"reason,omitempty"`
		} `json:"validationErrors,omitempty"`
	} `json:"extensions,omitempty"`
}

// GetQueryParam represents a query parameter to used with GraphQL GET requests.
type GetQueryParam struct {
	Query     string `url:"query,omitempty"`
	Variables string `url:"variables,omitempty"`
}

func (r *ErrorResponse) Error() string {
	if len(r.Errors) > 0 {
		var messages []string
		for _, e := range r.Errors {
			f, _ := json.Marshal(e.DownstreamResponse)

			messages = append(messages, e.Message)
			messages = append(messages, string(f))
		}
		return strings.Join(messages, ", ")
	}

	return ""
}
