package models

// Response is the standard API response wrapper
type Response struct {
	Data interface{} `json:"data"`           // Actual response data (can be any type)
	Meta *Meta       `json:"meta,omitempty"` // Optional pagination metadata
}

// Meta contains pagination information
type Meta struct {
	Total  int `json:"total,omitempty"`  // Total number of items available
	Limit  int `json:"limit,omitempty"`  // Maximum items returned in this response
	Offset int `json:"offset,omitempty"` // Starting position in the full result set
}

// ErrorResponse is returned when an API error occurs
type ErrorResponse struct {
	Error   string `json:"error"`             // Human-readable error message
	Code    string `json:"code,omitempty"`    // Machine-readable error code
	Details string `json:"details,omitempty"` // Additional error details
}
