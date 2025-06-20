package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	// "log" // Uncomment if logging within writeJSONResponse's json.NewEncoder error

	"app/database"
	"app/models"

	"github.com/go-chi/chi/v5" // Added for URLParam
	// "github.com/gorilla/mux" // Replaced by chi
)

// writeJSONResponse is a helper function to write JSON responses.
// If the status code indicates an error (>= 400), and the payload is a string or error,
// it formats the response as {"error": "message"}.
func writeJSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	responsePayload := payload
	if statusCode >= 400 {
		if errMessage, ok := payload.(string); ok {
			responsePayload = map[string]string{"error": errMessage}
		} else if err, ok := payload.(error); ok {
			responsePayload = map[string]string{"error": err.Error()}
		}
		// If payload is neither string nor error (e.g., already a map/struct for specific error details),
		// it will be passed as is. This allows flexibility.
	}

	w.WriteHeader(statusCode) // Set status code *after* potential payload modification

	if responsePayload != nil {
		if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
			// Log this error, as it's happening after headers are written.
			// http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError) // This would fail as headers are sent
			// Consider logging this failure internally, e.g., log.Printf("Error encoding JSON for status %d: %v", statusCode, err)
		}
	}
}

// GetItemsHandler handles retrieving all items.
func GetItemsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := database.GetItems(db)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve items")
			return
		}
		if items == nil {
			items = []models.Item{}
		}
		writeJSONResponse(w, http.StatusOK, items)
	}
}

// GetItemHandler was here (now handled by OpenAPI spec)
