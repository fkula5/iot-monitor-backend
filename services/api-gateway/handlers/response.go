package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/skni-kod/iot-monitor-backend/pkg/logger"
	"go.uber.org/zap"
)

// Response is the standard API response structure
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// JSON sends a standard JSON response
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode JSON response", zap.Error(err))
		// We can't use Error() here as headers are already sent
	}
}

// Error sends a standard JSON error response
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Error: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode JSON error response", zap.Error(err))
	}
}

// Created sends a 201 Created response with data
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
