package openpoiapi

import "fmt"

// APIError is returned when the API responds with a non-2xx status code.
// Use errors.As to inspect it, for example to detect rate limiting (429).
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Message is the "error" or "message" field of the response body, or the
	// standard status text when the body has neither.
	Message string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("openpoiapi: status %d: %s", e.StatusCode, e.Message)
}

// ValidationError is returned when request parameters fail client-side
// validation. No HTTP request is sent in that case.
type ValidationError struct {
	// Field is the invalid parameter: "center", "bbox" or "radius".
	Field string
	// Reason describes why the parameter is invalid.
	Reason string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("openpoiapi: invalid %s: %s", e.Field, e.Reason)
}
