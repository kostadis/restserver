package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	stdruntime "runtime" // Standard runtime
	"strconv"
	"strings"
	"testing"

	"app/database"
	"app/internal/generated/openapi" // Added for generated types & its local error types
	"app/models"                     // Original model, still used for creating test data

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Apply schema
	schemaPath := filepath.Join(getProjectRootForHandlers(), "database", "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "Failed to read schema.sql")
	_, err = db.Exec(string(schemaBytes))
	require.NoError(t, err, "Failed to execute schema on test database")

	return db
}

// setupTestRouter initializes a Chi router with the necessary handlers for testing.
// This version is updated to only use OpenAPI handlers for Create, GetByID, and UpdateByID.
func setupTestRouter(db *sql.DB) *chi.Mux {
	router := chi.NewRouter()
	itemAPIServer := NewItemAPIServer(db)

	// Register OpenAPI handlers (includes POST /items, GET /items/{id}, PUT /items/{id})
	openapi.HandlerWithOptions(itemAPIServer, openapi.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			status := http.StatusBadRequest // Default
			var e *openapi.InvalidParamFormatError
			if errors.As(err, &e) {
				status = http.StatusBadRequest
			} else if strings.Contains(err.Error(), "found") { // Simple check for "not found" type errors
				status = http.StatusNotFound
			}
			// Add more specific error type checks if needed from oapi-codegen/runtime
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(openapi.Error{Error: err.Error()})
		},
	})

	// Register other non-OpenAPI routes (if any) that are still managed by old handlers
	router.Get("/items", GetItemsHandler(db)) // For getting all items (assuming this is not OpenAPI yet)
	// router.Put("/items/{id}", UpdateItemHandler(db)) // REMOVED - Now handled by OpenAPI
	router.Delete("/items/{id}", DeleteItemHandler(db)) // Assuming this is not OpenAPI yet

	return router
}

// PtrString is a helper function to get a pointer to a string.
func PtrString(s string) *string {
	return &s
}

// Helper to create an item directly in the DB for test setup
func createTestItemDirectly(t *testing.T, db *sql.DB, item models.Item) models.Item {
	id, err := database.CreateItem(db, item)
	require.NoError(t, err)
	item.ID = id
	return item
}

func TestCreateItemOpenAPI(t *testing.T) {
	db := setupHandlerTestDB(t) // This already applies schema
	defer db.Close()

	itemAPIServer := NewItemAPIServer(db)
	router := chi.NewRouter()
	openapi.HandlerWithOptions(itemAPIServer, openapi.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(openapi.Error{Error: "test error handler: " + err.Error()})
		},
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("Successful item creation via OpenAPI", func(t *testing.T) {
		newItem := openapi.NewItem{
			Name:        "Test Item via OpenAPI",
			Priority:    10,
			Description: PtrString("A description for OpenAPI test"),
		}
		bodyBytes, _ := json.Marshal(newItem)
		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)
		var createdItem openapi.Item
		err = json.NewDecoder(res.Body).Decode(&createdItem)
		require.NoError(t, err)
		assert.Equal(t, newItem.Name, createdItem.Name)
		assert.Equal(t, newItem.Priority, createdItem.Priority)
		assert.EqualValues(t, newItem.Description, createdItem.Description)
		require.NotNil(t, createdItem.Id)
		assert.True(t, *createdItem.Id > 0)
	})

	t.Run("Bad request via OpenAPI - missing name", func(t *testing.T) {
		badItem := openapi.NewItem{Priority: 5, Description: PtrString("Item with no name")}
		bodyBytes, _ := json.Marshal(badItem)
		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		var errResp openapi.Error
		_ = json.NewDecoder(res.Body).Decode(&errResp)
		assert.Contains(t, errResp.Error, "Name is required")
	})

	t.Run("Bad request via OpenAPI - invalid priority", func(t *testing.T) {
		badItem := openapi.NewItem{Name: "Item with bad priority", Priority: 0, Description: PtrString("A description")}
		bodyBytes, _ := json.Marshal(badItem)
		res, err := http.Post(ts.URL+"/items", "application/json", bytes.NewBuffer(bodyBytes))
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		var errResp openapi.Error
		_ = json.NewDecoder(res.Body).Decode(&errResp)
		assert.Contains(t, errResp.Error, "Priority must be a positive integer")
	})

	t.Run("Bad request via OpenAPI - malformed JSON", func(t *testing.T) {
		malformedJSON := `{"name": "Test", "priority": 1, "description": "Test desc"`
		res, err := http.Post(ts.URL+"/items", "application/json", strings.NewReader(malformedJSON))
		require.NoError(t, err)
		defer res.Body.Close()
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
		var errResp openapi.Error
		_ = json.NewDecoder(res.Body).Decode(&errResp)
		assert.Contains(t, errResp.Error, "Invalid request payload")
	})
}

func TestGetItemsHandler(t *testing.T) { // Assuming this remains non-OpenAPI for now
	db := setupHandlerTestDB(t)
	defer db.Close()
	createTestItemDirectly(t, db, models.Item{Name: "Item1", Description: "Desc1", Priority: 1})
	createTestItemDirectly(t, db, models.Item{Name: "Item2", Description: "Desc2", Priority: 2})
	router := setupTestRouter(db)

	req, _ := http.NewRequest(http.MethodGet, "/items", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var items []models.Item
	err := json.NewDecoder(rr.Body).Decode(&items)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetItemByIdOpenAPI(t *testing.T) { // Renamed to avoid conflict if an old GetItemByIdHandler test existed
	db := setupHandlerTestDB(t)
	defer db.Close()
	initialItem := createTestItemDirectly(t, db, models.Item{Name: "SpecificItem", Description: "Specific Description", Priority: 3})
	router := setupTestRouter(db) // This router uses OpenAPI for /items/{id} GET

	t.Run("found", func(t *testing.T) {
		reqPath := "/items/" + strconv.FormatInt(initialItem.ID, 10)
		req, _ := http.NewRequest(http.MethodGet, reqPath, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var item openapi.Item
		err := json.NewDecoder(rr.Body).Decode(&item)
		require.NoError(t, err)
		require.NotNil(t, item.Id)
		assert.Equal(t, initialItem.ID, *item.Id)
		assert.Equal(t, initialItem.Name, item.Name)
		require.NotNil(t, item.Description)
		assert.Equal(t, initialItem.Description, *item.Description)
		assert.Equal(t, int32(initialItem.Priority), item.Priority)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/items/99999", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		var errResp openapi.Error
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Item not found")
	})

	t.Run("invalid id format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/items/abc", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code) // Error handler in setupTestRouter should catch this
		var errResp openapi.Error
		err := json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, strings.ToLower(errResp.Error), "invalid format for parameter id")
	})
}

func TestUpdateItemOpenAPI(t *testing.T) {
	db := setupHandlerTestDB(t) // This also applies the schema
	defer db.Close()

	router := setupTestRouter(db) // This router includes the OpenAPI PUT handler
	ts := httptest.NewServer(router)
	defer ts.Close()

	// Pre-populate an item to update
	initialItemModel := createTestItemDirectly(t, db, models.Item{Name: "Initial Item", Priority: 1, Description: "Initial Description"})

	client := &http.Client{}

	t.Run("Successful Update", func(t *testing.T) {
		updatePayload := openapi.UpdateItem{
			Name:        "Updated Item Name",
			Priority:    5,
			Description: PtrString("Updated Description"),
		}
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, initialItemModel.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedAPIItem openapi.Item
		err = json.NewDecoder(resp.Body).Decode(&updatedAPIItem)
		require.NoError(t, err)

		assert.Equal(t, updatePayload.Name, updatedAPIItem.Name)
		assert.Equal(t, updatePayload.Priority, updatedAPIItem.Priority)
		require.NotNil(t, updatedAPIItem.Description)
		assert.Equal(t, *updatePayload.Description, *updatedAPIItem.Description)
		require.NotNil(t, updatedAPIItem.Id)
		assert.Equal(t, initialItemModel.ID, *updatedAPIItem.Id)

		// Verify in DB
		dbItem, err := database.GetItem(db, initialItemModel.ID)
		require.NoError(t, err)
		assert.Equal(t, updatePayload.Name, dbItem.Name)
		assert.Equal(t, int(updatePayload.Priority), dbItem.Priority)
		assert.Equal(t, *updatePayload.Description, dbItem.Description)
	})

	t.Run("Successful Update with nil description", func(t *testing.T) {
		// Create a new item for this specific test case to avoid interference
		itemToUpdate := createTestItemDirectly(t, db, models.Item{Name: "Item For Nil Desc Update", Priority: 2, Description: "Existing Description"})

		updatePayload := openapi.UpdateItem{
			Name:        "Updated Name For Nil Desc",
			Priority:    int32(itemToUpdate.Priority), // Keep priority same or change, doesn't matter much for this test
			Description: nil,                   // Explicitly set description to nil
		}
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, itemToUpdate.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respItem openapi.Item
		err = json.NewDecoder(resp.Body).Decode(&respItem)
		require.NoError(t, err)

		assert.Equal(t, updatePayload.Name, respItem.Name)
		// In the handler, if requestBody.Description is nil, dbItem.Description becomes ""
		// Then, when fetching for response, responseItem.Description becomes &""
		require.NotNil(t, respItem.Description, "Description should be non-nil pointer to empty string")
		assert.Equal(t, "", *respItem.Description, "Description in response should be empty string")


		// Verify in DB
		dbItem, err := database.GetItem(db, itemToUpdate.ID)
		require.NoError(t, err)
		assert.Equal(t, "", dbItem.Description, "Description in DB should be empty string")
	})

	t.Run("Successful Update with empty string description", func(t *testing.T) {
		itemToUpdate := createTestItemDirectly(t, db, models.Item{Name: "Item For Empty Desc Update", Priority: 3, Description: "Non-empty description"})

		updatePayload := openapi.UpdateItem{
			Name:        "Updated Name For Empty Desc",
			Priority:    int32(itemToUpdate.Priority),
			Description: PtrString(""), // Explicitly set description to pointer to empty string
		}
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, itemToUpdate.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respItem openapi.Item
		err = json.NewDecoder(resp.Body).Decode(&respItem)
		require.NoError(t, err)

		assert.Equal(t, updatePayload.Name, respItem.Name)
		require.NotNil(t, respItem.Description)
		assert.Equal(t, "", *respItem.Description)

		// Verify in DB
		dbItem, err := database.GetItem(db, itemToUpdate.ID)
		require.NoError(t, err)
		assert.Equal(t, "", dbItem.Description)
	})


	t.Run("Item Not Found (404)", func(t *testing.T) {
		updatePayload := openapi.UpdateItem{Name: "Any Name", Priority: 1}
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/999999", ts.URL) // Non-existent ID
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		var errResp openapi.Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Item not found")
	})

	t.Run("Invalid Payload - Missing Name (400)", func(t *testing.T) {
		updatePayload := openapi.UpdateItem{Priority: 1} // Name is missing
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, initialItemModel.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var errResp openapi.Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Name is required")
	})

	t.Run("Invalid Payload - Invalid Priority (400)", func(t *testing.T) {
		updatePayload := openapi.UpdateItem{Name: "Test Name", Priority: 0} // Invalid priority
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, initialItemModel.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var errResp openapi.Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Priority must be a positive integer")
	})

	t.Run("Invalid Item ID in Path (not an integer)", func(t *testing.T) {
		updatePayload := openapi.UpdateItem{Name: "Any Name", Priority: 1}
		payloadBytes, _ := json.Marshal(updatePayload)
		reqURL := fmt.Sprintf("%s/items/notaninteger", ts.URL)
		req, _ := http.NewRequest(http.MethodPut, reqURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// This error is caught by the custom ErrorHandlerFunc in setupTestRouter,
		// which wraps the oapi-codegen runtime's parameter binding error.
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var errResp openapi.Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, strings.ToLower(errResp.Error), "invalid format for parameter id")
	})

	t.Run("Malformed JSON payload", func(t *testing.T) {
		malformedJSON := `{"name": "Test", "priority": 1, "description": "Test desc"` // Missing closing brace
		reqURL := fmt.Sprintf("%s/items/%d", ts.URL, initialItemModel.ID)
		req, _ := http.NewRequest(http.MethodPut, reqURL, strings.NewReader(malformedJSON))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var errResp openapi.Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp.Error, "Invalid request payload")
	})
}

func TestDeleteItemHandler(t *testing.T) { // Assuming this remains non-OpenAPI for now
	db := setupHandlerTestDB(t)
	defer db.Close()
	initialItem := createTestItemDirectly(t, db, models.Item{Name: "ToDelete", Description: "Delete Desc", Priority: 1})
	router := setupTestRouter(db)

	reqPath := "/items/" + strconv.FormatInt(initialItem.ID, 10)
	req, _ := http.NewRequest(http.MethodDelete, reqPath, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	_, err := database.GetItem(db, initialItem.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows), "Expected sql.ErrNoRows after deleting item")
}

// TestCreateItemHandler (old, non-OpenAPI one) is removed as POST /items is covered by TestCreateItemOpenAPI
// TestUpdateItemHandler (old, non-OpenAPI one) is removed as PUT /items/{id} is covered by TestUpdateItemOpenAPI

// Test for GetItemByIdHandler is renamed to TestGetItemByIdOpenAPI and uses the OpenAPI router.
// Test for CreateItemHandler is renamed to TestCreateItemOpenAPI and uses the OpenAPI router.

// TestCreateItemHandler_BadRequest is removed as its cases are now covered by TestCreateItemOpenAPI (missing name, malformed JSON)
// or similar tests for TestUpdateItemOpenAPI.
// The original TestCreateItemHandler_BadRequest had specific checks for "Invalid request payload" vs "Name is required".
// The OpenAPI handlers have similar logic, so those specific cases are tested within TestCreateItemOpenAPI and TestUpdateItemOpenAPI.

// Note: The old TestCreateItemHandler is removed because POST /items is now served by openapi_handlers.CreateItem.
// Its functionality is tested in TestCreateItemOpenAPI.

// Note: The old TestGetItemByIdHandler has been effectively replaced/updated by TestGetItemByIdOpenAPI
// which uses the setupTestRouter that correctly routes GET /items/{id} to the openapi_handlers.GetItemById.

// Note: The old TestCreateItemHandler_BadRequest has been removed. Its cases (malformed JSON, missing fields)
// are covered by equivalent test cases within TestCreateItemOpenAPI and TestUpdateItemOpenAPI.
// The new handlers provide similar error messages for these scenarios.
