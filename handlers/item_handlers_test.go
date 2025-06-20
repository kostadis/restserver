package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	stdruntime "runtime" // Standard runtime
	"strconv"
	"strings"
	"testing"
	"errors" // For custom error handler
	"os"     // For os.ReadFile

	"app/database"
	"app/models" // Original model, still used for creating test data
	"app/internal/generated/openapi" // Added for generated types & its local error types

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// getProjectRootForHandlers uses standard runtime.
func getProjectRootForHandlers() string {
	_, b, _, _ := stdruntime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..")
}

// setupHandlerTestDB remains the same.
func setupHandlerTestDB(t *testing.T) *sql.DB {
	db, err := database.InitDB(":memory:")
	require.NoError(t, err, "Failed to initialize test database for handlers")
	return db
}

// setupTestRouter initializes a Chi router with the necessary handlers for testing.
func setupTestRouter(db *sql.DB) *chi.Mux {
	router := chi.NewRouter()
	itemAPIServer := NewItemAPIServer(db)

	openapi.HandlerWithOptions(itemAPIServer, openapi.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			var status int
			var e *openapi.InvalidParamFormatError // Use error type from the generated openapi package
			if errors.As(err, &e) {
				status = http.StatusBadRequest
			} else {
				status = http.StatusBadRequest
			}
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(openapi.Error{Error: err.Error()})
		},
	})

	// router.Post("/items", CreateItemHandler(db)) // This handler was removed and is covered by OpenAPI
	router.Get("/items", GetItemsHandler(db))
	router.Put("/items/{id}", UpdateItemHandler(db))
	router.Delete("/items/{id}", DeleteItemHandler(db))

	return router
}


func TestCreateItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	router := setupTestRouter(db)

	itemPayload := models.Item{
		Name:        "Test Handler Item",
		Description: "Description for handler test",
		Priority:    1,
	}
	payloadBytes, err := json.Marshal(itemPayload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var createdItem models.Item
	err = json.NewDecoder(rr.Body).Decode(&createdItem)
	require.NoError(t, err)
	assert.NotZero(t, createdItem.ID)
	assert.Equal(t, itemPayload.Name, createdItem.Name)
}

func TestGetItemsHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	_, err := database.CreateItem(db, models.Item{Name: "Item1", Description: "Desc1", Priority: 1})
	require.NoError(t, err)
	_, err = database.CreateItem(db, models.Item{Name: "Item2", Description: "Desc2", Priority: 2})
	require.NoError(t, err)
	router := setupTestRouter(db)

	req, err := http.NewRequest(http.MethodGet, "/items", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var items []models.Item
	err = json.NewDecoder(rr.Body).Decode(&items)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetItemByIdHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	createdItemInput := models.Item{Name: "SpecificItem", Description: "Specific Description", Priority: 3}
	createdID, err := database.CreateItem(db, createdItemInput)
	require.NoError(t, err)
	router := setupTestRouter(db)

	t.Run("found", func(t *testing.T) {
		reqPath := "/items/" + strconv.FormatInt(createdID, 10)
		req, err := http.NewRequest(http.MethodGet, reqPath, nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var item openapi.Item
		err = json.NewDecoder(rr.Body).Decode(&item)
		require.NoError(t, err)
		require.NotNil(t, item.Id)
		assert.Equal(t, createdID, *item.Id)
		assert.Equal(t, createdItemInput.Name, item.Name)
		require.NotNil(t, item.Description)
		assert.Equal(t, createdItemInput.Description, *item.Description)
		assert.Equal(t, int32(createdItemInput.Priority), item.Priority)
	})

	t.Run("not found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/items/99999", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		var errResp openapi.Error
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Item not found")
	})

	t.Run("invalid id format", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/items/abc", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp openapi.Error
		decodeErr := json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, decodeErr, "Failed to decode error response into openapi.Error for invalid id format")
		assert.NotEmpty(t, errResp.Error)
		assert.Contains(t, strings.ToLower(errResp.Error), "invalid format for parameter", "Error message should indicate invalid format for parameter")
	})
}

// PtrString is a helper function to get a pointer to a string.
// Useful for optional string fields in OpenAPI generated structs.
func PtrString(s string) *string {
	return &s
}

func TestCreateItemOpenAPI(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	// Read and execute schema
	// Adjust path if getProjectRootForHandlers() is not suitable or if tests are run from a different CWD.
	// Assuming tests are run from the 'handlers' directory or project root where '../database/schema.sql' is valid.
	schemaPath := filepath.Join(getProjectRootForHandlers(), "database", "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read schema.sql")
	_, err = db.Exec(string(schemaBytes))
	require.NoError(t, err, "Failed to execute schema on test database")

	itemAPIServer := NewItemAPIServer(db) // Use the actual ItemAPIServer
	router := chi.NewRouter()          // Create a new router for this specific test setup

	// Register ONLY the OpenAPI handlers. This ensures that POST /items uses the openapi_handlers.CreateItem
	openapi.HandlerWithOptions(itemAPIServer, openapi.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) { // Optional: Custom error handler for tests
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest) // Default to 400 for test simplicity
			json.NewEncoder(w).Encode(openapi.Error{Error: "test error handler: " + err.Error()})
		},
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("Successful item creation via OpenAPI", func(t *testing.T) {
		newItem := openapi.NewItem{
			Name:        "Test Item via OpenAPI",
			Priority:    10,
			Description: PtrString("A description for OpenAPI test"), // Use local PtrString
		}
		bodyBytes, err := json.Marshal(newItem)
		require.NoError(t, err)

		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusCreated, res.StatusCode, "Expected status 201 Created")

		var createdItem openapi.Item
		err = json.NewDecoder(res.Body).Decode(&createdItem)
		require.NoError(t, err, "Failed to decode successful response")

		assert.Equal(t, newItem.Name, createdItem.Name)
		assert.Equal(t, newItem.Priority, createdItem.Priority)
		require.NotNil(t, newItem.Description, "Test setup error: newItem.Description should not be nil")
		require.NotNil(t, createdItem.Description, "createdItem.Description should not be nil for this test case")
		if createdItem.Description != nil && newItem.Description != nil { // Defensive check
			assert.Equal(t, *newItem.Description, *createdItem.Description)
		}
		require.NotNil(t, createdItem.Id, "Created item ID should not be nil")
		assert.True(t, *createdItem.Id > 0, "Created item ID should be positive")
	})

	t.Run("Bad request via OpenAPI - missing name", func(t *testing.T) {
		badItem := openapi.NewItem{
			// Name is intentionally missing
			Priority:    5,
			Description: PtrString("Item with no name"), // Use local PtrString
		}
		bodyBytes, err := json.Marshal(badItem)
		require.NoError(t, err)

		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode, "Expected status 400 Bad Request")

		var errResp openapi.Error
		err = json.NewDecoder(res.Body).Decode(&errResp)
		require.NoError(t, err, "Failed to decode error response")

		// The error message comes from the validation in handlers/openapi_handlers.go CreateItem
		expectedErrorMsg := "Name is required"
		assert.Contains(t, errResp.Error, expectedErrorMsg, "Error message mismatch")
	})

	t.Run("Bad request via OpenAPI - invalid priority", func(t *testing.T) {
		badItem := openapi.NewItem{
			Name:        "Item with bad priority",
			Priority:    0, // Priority must be positive
			Description: PtrString("A description"),
		}
		bodyBytes, err := json.Marshal(badItem)
		require.NoError(t, err)

		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		var errResp openapi.Error
		err = json.NewDecoder(res.Body).Decode(&errResp)
		require.NoError(t, err)
		expectedErrorMsg := "Priority must be a positive integer"
		assert.Contains(t, errResp.Error, expectedErrorMsg)
	})

	t.Run("Bad request via OpenAPI - malformed JSON", func(t *testing.T) {
		malformedJSON := `{"name": "Test", "priority": 1, "description": "Test desc"` // Missing closing brace

		res, err := http.Post(ts.URL+"/items", "application/json", strings.NewReader(malformedJSON))
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		var errResp openapi.Error
		err = json.NewDecoder(res.Body).Decode(&errResp)
		require.NoError(t, err, "Failed to decode malformed JSON error response")
		// Error comes from json.NewDecoder in the handler
		assert.Contains(t, errResp.Error, "Invalid request payload", "Expected error message for malformed JSON")
	})
}

func TestUpdateItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	initialItem := models.Item{Name: "InitialName", Description: "Initial Desc", Priority: 1}
	initialID, err := database.CreateItem(db, initialItem)
	require.NoError(t, err)
	router := setupTestRouter(db)

	updatePayload := models.Item{Name: "UpdatedName", Description: "Updated Desc", Priority: 2}
	payloadBytes, _ := json.Marshal(updatePayload)
	reqPath := "/items/" + strconv.FormatInt(initialID, 10)
	req, err := http.NewRequest(http.MethodPut, reqPath, bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var updatedItem models.Item
	err = json.NewDecoder(rr.Body).Decode(&updatedItem)
	require.NoError(t, err)
	assert.Equal(t, "UpdatedName", updatedItem.Name)
	assert.Equal(t, initialID, updatedItem.ID)
}

func TestDeleteItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	itemToDelete := models.Item{Name: "ToDelete", Description: "Delete Desc", Priority: 1}
	initialID, err := database.CreateItem(db, itemToDelete)
	require.NoError(t, err)
	router := setupTestRouter(db)

	reqPath := "/items/" + strconv.FormatInt(initialID, 10)
	req, err := http.NewRequest(http.MethodDelete, reqPath, nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	_, err = database.GetItem(db, initialID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err, "Expected sql.ErrNoRows after deleting item")
}

func TestCreateItemHandler_BadRequest(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()
	router := setupTestRouter(db)

	t.Run("malformed json", func(t *testing.T) {
		// Malformed: missing closing brace for name string, and overall closing brace
		req, err := http.NewRequest(http.MethodPost, "/items", strings.NewReader("{\"name\": \"bad json value"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp map[string]string
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp["error"], "Invalid request payload", "Expected 'Invalid request payload' for malformed JSON")
	})

	t.Run("missing required fields", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/items", strings.NewReader("{\"description\": \"only description\"}"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var errResp map[string]string
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		// This test now hits the OpenAPI handler which has a more specific error message
		// when only name is missing from a JSON that parses to openapi.NewItem.
		// The openapi.NewItem struct has Name as a required field (non-pointer string),
		// so if the JSON is `{"priority": 1}`, Name defaults to "".
		// The handler then checks `if requestBody.Name == ""`
		assert.Contains(t, errResp["error"], "Name is required", "Expected validation error for missing name via OpenAPI handler")
	})
}
