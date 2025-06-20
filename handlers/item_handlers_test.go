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

	router.Post("/items", CreateItemHandler(db))
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
		assert.Contains(t, errResp["error"], "Name and a positive Priority are required", "Expected validation error for missing fields")
	})
}
