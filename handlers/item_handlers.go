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

	"github.com/gorilla/mux"
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

// CreateItemHandler handles the creation of a new item.
func CreateItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var item models.Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}
		defer r.Body.Close()

		if item.Name == "" || item.Priority <= 0 {
			writeJSONResponse(w, http.StatusBadRequest, "Name and a positive Priority are required")
			return
		}

		id, err := database.CreateItem(db, item)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, "Failed to create item") // User-friendly message
			return
		}
		item.ID = id
		writeJSONResponse(w, http.StatusCreated, item)
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

// GetItemHandler handles retrieving a single item by its ID.
func GetItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			// This case should ideally not happen if route is defined correctly
			writeJSONResponse(w, http.StatusBadRequest, "Item ID not provided in URL path")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, "Invalid item ID format")
			return
		}

		item, err := database.GetItem(db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONResponse(w, http.StatusNotFound, "Item not found")
			} else {
				writeJSONResponse(w, http.StatusInternalServerError, "Failed to retrieve item")
			}
			return
		}
		writeJSONResponse(w, http.StatusOK, item)
	}
}

// UpdateItemHandler handles updating an existing item.
func UpdateItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			writeJSONResponse(w, http.StatusBadRequest, "Item ID not provided in URL path")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, "Invalid item ID format")
			return
		}

		var item models.Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}
		defer r.Body.Close()

		if item.Name == "" || item.Priority <= 0 {
			writeJSONResponse(w, http.StatusBadRequest, "Name and a positive Priority are required")
			return
		}

		item.ID = id // Ensure ID from path is used

		rowsAffected, err := database.UpdateItem(db, id, item)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONResponse(w, http.StatusNotFound, "Item not found to update")
			} else {
				writeJSONResponse(w, http.StatusInternalServerError, "Failed to update item")
			}
			return
		}

		if rowsAffected == 0 {
			writeJSONResponse(w, http.StatusNotFound, "Item not found, no update occurred")
			return
		}

		updatedItem, err := database.GetItem(db, id)
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, "Item updated, but failed to retrieve confirmation")
			return
		}
		writeJSONResponse(w, http.StatusOK, updatedItem)
	}
}

// DeleteItemHandler handles deleting an item by its ID.
func DeleteItemHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr, ok := vars["id"]
		if !ok {
			writeJSONResponse(w, http.StatusBadRequest, "Item ID not provided in URL path")
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, "Invalid item ID format")
			return
		}

		rowsAffected, err := database.DeleteItem(db, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONResponse(w, http.StatusNotFound, "Item not found to delete")
			} else {
				writeJSONResponse(w, http.StatusInternalServerError, "Failed to delete item")
			}
			return
		}

		if rowsAffected == 0 {
			writeJSONResponse(w, http.StatusNotFound, "Item not found, no deletion occurred")
			return
		}
		writeJSONResponse(w, http.StatusNoContent, nil)
	}
}
