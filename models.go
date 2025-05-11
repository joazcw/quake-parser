package main

// ErrorResponse represents the structure of error responses returned by the API.
// This is primarily used for Swagger documentation.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response with a message.
// This is primarily used for Swagger documentation.
type SuccessResponse struct {
	Message string `json:"message"`
}

// UploadResponse represents the response for a file upload operation.
// This is primarily used for Swagger documentation.
type UploadResponse struct {
	Message        string `json:"message"`
	GamesProcessed int    `json:"games_processed"`
} 