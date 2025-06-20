package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"app/database"
	// "app/models" // Original model - This was unused, GetItem now returns database.Item which is mapped to openapi.Item
	"app/internal/generated/openapi" // Generated package
)

// ItemAPIServer implements the openapi.ServerInterface
type ItemAPIServer struct {
	DB *sql.DB
}

// Ensure ItemAPIServer implements the interface.
var _ openapi.ServerInterface = (*ItemAPIServer)(nil)

// GetItemById implements the logic for the (GET /items/{id}) endpoint.
func (s *ItemAPIServer) GetItemById(w http.ResponseWriter, r *http.Request, id int64) {
	dbItem, err := database.GetItem(s.DB, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(openapi.Error{Error: "Item not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(openapi.Error{Error: "Failed to retrieve item: " + err.Error()})
		}
		return
	}

	// Convert database.Item to openapi.Item
	apiItem := openapi.Item{
		Id:          &dbItem.ID, // OpenAPI schema has ID as pointer
		Name:        dbItem.Name,
		Description: &dbItem.Description, // OpenAPI schema has Description as pointer
		Priority:    int32(dbItem.Priority), // OpenAPI schema has Priority as int32
	}
	// Handle cases where Description might be empty an we don't want to send null but omit the field if possible
	// The current openapi.Item struct uses *string, so an empty string will be sent as ""
	// If dbItem.Description is empty, apiItem.Description will be a pointer to an empty string.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiItem); err != nil {
		// Log error internally, as headers are already written
		// For now, we can't send another HTTP error to the client here.
		http.Error(w, "Failed to write response", http.StatusInternalServerError) // This might not reach client if headers sent.
	}
}

// NewItemAPIServer creates a new ItemAPIServer.
func NewItemAPIServer(db *sql.DB) *ItemAPIServer {
	return &ItemAPIServer{DB: db}
}
