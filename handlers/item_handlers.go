package handlers

import (
	"database/sql"
	// "encoding/json" // No longer needed after removing writeJSONResponse and GetItemsHandler
	"errors"
	"net/http"
	// "strconv" // No longer needed
	// "log" // Uncomment if logging within writeJSONResponse's json.NewEncoder error

	"app/database"
	"app/models"

	// "github.com/go-chi/chi/v5" // No longer needed
	// "github.com/gorilla/mux" // Replaced by chi
)

// GetItemHandler was here (now handled by OpenAPI spec)
// GetItemsHandler was here (now handled by OpenAPI spec)
// writeJSONResponse was here (no longer needed as its only user GetItemsHandler was removed)
