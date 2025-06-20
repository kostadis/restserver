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
// This line will cause a compile error if the interface is not fully implemented.
var _ openapi.ServerInterface = (*ItemAPIServer)(nil)

// NewItemAPIServer creates a new ItemAPIServer.
func NewItemAPIServer(db *sql.DB) *ItemAPIServer {
	return &ItemAPIServer{DB: db}
}

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
		Id:          &dbItem.ID,
		Name:        dbItem.Name,
		Description: &dbItem.Description,
		Priority:    int32(dbItem.Priority),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(apiItem); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
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

	dbItem := models.Item{
		Name:     requestBody.Name,
		Priority: int(requestBody.Priority),
	}
	if requestBody.Description != nil {
		dbItem.Description = *requestBody.Description
	} else {
		dbItem.Description = "" // Default to empty string if not provided
	}

	id, err := database.CreateItem(s.DB, dbItem)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(openapi.Error{Error: "Failed to create item: " + err.Error()})
		return
	}

	responseItem := openapi.Item{
		Id:          &id,
		Name:        requestBody.Name,
		Priority:    requestBody.Priority,
		Description: requestBody.Description, // Pass through the *string
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(responseItem); err != nil {
		// Log error
	}
}

// UpdateItemById implements the logic for the (PUT /items/{id}) endpoint.
func (s *ItemAPIServer) UpdateItemById(w http.ResponseWriter, r *http.Request, id int64) {
    var requestBody openapi.UpdateItem // Generated struct for the request body
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(openapi.Error{Error: "Invalid request payload: " + err.Error()})
        return
    }
    defer r.Body.Close()

    if requestBody.Name == "" { // Name is required by schema, but explicit check is good
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(openapi.Error{Error: "Name is required"})
        return
    }
    if requestBody.Priority <= 0 { // Priority is required and must be positive
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(openapi.Error{Error: "Priority must be a positive integer"})
        return
    }

    dbItem := models.Item{
        ID:       id, // ID from path parameter
        Name:     requestBody.Name,
        Priority: int(requestBody.Priority), // Convert int32 to int
    }
    if requestBody.Description != nil {
        dbItem.Description = *requestBody.Description
    } else {
        dbItem.Description = "" // Assuming models.Item.Description is string and not nullable in DB
    }

    rowsAffected, err := database.UpdateItem(s.DB, id, dbItem)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        if errors.Is(err, sql.ErrNoRows) {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(openapi.Error{Error: "Item not found to update"})
        } else {
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(openapi.Error{Error: "Failed to update item: " + err.Error()})
        }
        return
    }

    if rowsAffected == 0 { // Should ideally be covered by sql.ErrNoRows from UpdateItem
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(openapi.Error{Error: "Item not found, or no changes made"})
        return
    }

    updatedDbItem, err := database.GetItem(s.DB, id)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(openapi.Error{Error: "Item updated, but failed to retrieve confirmation: " + err.Error()})
        return
    }

    responseItem := openapi.Item{
        Id:          &updatedDbItem.ID,
        Name:        updatedDbItem.Name,
        Description: &updatedDbItem.Description, // Convert string to *string for response
        Priority:    int32(updatedDbItem.Priority), // Convert int to int32 for response
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(responseItem); err != nil {
        // Log error, as headers are already written
    }
}
