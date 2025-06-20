package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	// "log" // Add if needed for debugging

	"app/database"
	"app/internal/generated/openapi" // Generated package
	"app/models"                     // For converting to DB model
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

// CreateItem handles the creation of a new item based on the OpenAPI spec.
func (s *ItemAPIServer) CreateItem(w http.ResponseWriter, r *http.Request) {
	var requestBody openapi.NewItem // This is the schema defined for the request body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(openapi.Error{Error: "Invalid request payload: " + err.Error()})
		return
	}
	defer r.Body.Close()

	// Validate input (example: Name and Priority are required by schema, but extra checks can be here)
	if requestBody.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(openapi.Error{Error: "Name is required"})
		return
	}
	if requestBody.Priority <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(openapi.Error{Error: "Priority must be a positive integer"})
		return
	}

	// Convert openapi.NewItem to models.Item for database operation
	dbItem := models.Item{
		Name:     requestBody.Name,
		Priority: int(requestBody.Priority), // models.Item uses int for Priority
	}
	if requestBody.Description != nil {
		dbItem.Description = *requestBody.Description
	}

	// Call database to create item
	id, err := database.CreateItem(s.DB, dbItem)
	if err != nil {
		// log.Printf("Error creating item in database: %v", err) // Optional logging
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(openapi.Error{Error: "Failed to create item: " + err.Error()})
		return
	}

	// Construct the response item (openapi.Item, which includes the ID)
	responseItem := openapi.Item{
		Id:          &id,
		Name:        requestBody.Name,
		Priority:    requestBody.Priority, // Retains int32 from NewItem/Item schema
		Description: requestBody.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responseItem); err != nil {
		// If encoding fails after setting headers, it's hard to recover gracefully.
		// Log the error. Consider if any other action is needed.
		// log.Printf("Error encoding success response: %v", err)
	}
}
