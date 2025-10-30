// Package python provides E2E test helpers for Python Lambda function testing.

package python

// OrderRequest represents the API request payload.
type OrderRequest struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// OrderResponse represents the API response payload.
type OrderResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Status string `json:"status"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}
