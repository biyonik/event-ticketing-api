package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}

// parseIDFromPath extracts ID from URL path
// Example: "/events/123" with prefix "/events/" returns 123
func parseIDFromPath(path, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.Split(idStr, "/")[0] // Handle paths like /events/123/publish

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid ID")
	}

	return id, nil
}

// getUserIDFromContext extracts user ID from request context
// In a real app, this would come from JWT middleware
func getUserIDFromContext(r *http.Request) int64 {
	// Placeholder: In real app, extract from JWT token
	return 1 // Hardcoded for demo
}
