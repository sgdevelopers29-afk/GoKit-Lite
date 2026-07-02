package response

import (
	"encoding/json"
	"net/http"
)

// Response is a struct that represents a standard API response format.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success creates a successful response with the given data.
func Success(data any) Response {
	return Response{
		Success: true,
		Message: "success",
		Data:    data,
	}
}

// Error creates an error response with the given message.
func Error(message string) Response {
	return Response{
		Success: false,
		Message: message,
		Data:    nil,
	}
}

// WriteJSON sends a standardized Response struct as a JSON HTTP response.
func WriteJSON(w http.ResponseWriter, statusCode int, r Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(r)
}

// WriteSuccess is a convenience function that sends a 200 OK JSON response with the given data.
func WriteSuccess(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, Success(data))
}

// WriteError is a convenience function that sends a JSON error response with the given status code and message.
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	WriteJSON(w, statusCode, Error(message))
}
