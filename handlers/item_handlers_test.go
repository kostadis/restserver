package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath" // For robust schema path finding for test DB
	"runtime"       // For robust schema path finding for test DB
	"strconv"
	"strings"
	"testing"

	"app/database" // To init a test DB and for handler dependencies
	"app/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// getProjectRootForHandlers returns the root directory of the project from handlers_test.go
func getProjectRootForHandlers() string {
	_, b, _, _ := runtime.Caller(0)
	// Root directory of project is one level up from handlers_test.go
	return filepath.Join(filepath.Dir(b), "..")
}

// setupHandlerTestDB initializes an in-memory SQLite database for handler tests.
// It's similar to the one in database_test.go but adapted for handler test context if needed.
// Crucially, it ensures that database.InitDB can find the schema.sql.
func setupHandlerTestDB(t *testing.T) *sql.DB {
	// The database.InitDB function has been updated to find schema.sql relative to its own package.
	// So, calling it with ":memory:" should now work correctly from any test file.
	db, err := database.InitDB(":memory:")
	require.NoError(t, err, "Failed to initialize test database for handlers")

	// Teardown (closing the DB) will be managed by the individual test functions
	// or a suite setup if using testify/suite. For individual tests, defer db.Close().
	return db
}

func TestCreateItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/items", CreateItemHandler(db)).Methods(http.MethodPost)

	itemPayload := models.Item{
		Name:        "Test Handler Item",
		Description: "Description for handler test",
		Price:       19.99,
	}
	payloadBytes, err := json.Marshal(itemPayload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/items", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req) // Use router to serve to match routes correctly

	assert.Equal(t, http.StatusCreated, rr.Code, "Handler returned wrong status code")

	var createdItem models.Item
	err = json.NewDecoder(rr.Body).Decode(&createdItem)
	require.NoError(t, err, "Failed to decode response body")

	assert.NotZero(t, createdItem.ID, "Expected created item to have an ID")
	assert.Equal(t, itemPayload.Name, createdItem.Name)
	assert.Equal(t, itemPayload.Description, createdItem.Description)
	assert.Equal(t, itemPayload.Price, createdItem.Price)

	// Verify in DB as well
	dbItem, err := database.GetItem(db, createdItem.ID)
	require.NoError(t, err, "Failed to get item from DB for verification")
	assert.Equal(t, createdItem.Name, dbItem.Name)
}

// TestGetItemsHandler outlines a test for getting all items.
func TestGetItemsHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	// Setup: Add some items to the DB
	_, err := database.CreateItem(db, models.Item{Name: "Item1", Price: 10})
	require.NoError(t, err)
	_, err = database.CreateItem(db, models.Item{Name: "Item2", Price: 20})
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/items", GetItemsHandler(db)).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/items", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var items []models.Item
	err = json.NewDecoder(rr.Body).Decode(&items)
	require.NoError(t, err)
	assert.Len(t, items, 2, "Expected two items")
}

// TestGetItemHandler outlines a test for getting a single item.
func TestGetItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	// Setup: Add an item
	created, err := database.CreateItem(db, models.Item{Name: "SpecificItem", Price: 12.34})
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/items/{id}", GetItemHandler(db)).Methods(http.MethodGet)

	t.Run("found", func(t *testing.T) {
		reqPath := "/items/" + strconv.FormatInt(created, 10)
		req, err := http.NewRequest(http.MethodGet, reqPath, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req) // router will match {id}

		assert.Equal(t, http.StatusOK, rr.Code)
		var item models.Item
		err = json.NewDecoder(rr.Body).Decode(&item)
		require.NoError(t, err)
		assert.Equal(t, "SpecificItem", item.Name)
	})

	t.Run("not found", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/items/9999", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		var errResp map[string]string
		err = json.NewDecoder(rr.Body).Decode(&errResp)
		require.NoError(t, err)
		assert.Contains(t, errResp["error"], "Item not found")
	})
}

// TestUpdateItemHandler outlines a test for updating an item.
func TestUpdateItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	initialID, err := database.CreateItem(db, models.Item{Name: "InitialName", Price: 1.00})
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/items/{id}", UpdateItemHandler(db)).Methods(http.MethodPut)

	updatePayload := models.Item{Name: "UpdatedName", Price: 2.00}
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
	assert.Equal(t, initialID, updatedItem.ID) // Ensure ID remains the same
}

// TestDeleteItemHandler outlines a test for deleting an item.
func TestDeleteItemHandler(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	initialID, err := database.CreateItem(db, models.Item{Name: "ToDelete", Price: 1.00})
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/items/{id}", DeleteItemHandler(db)).Methods(http.MethodDelete)

	reqPath := "/items/" + strconv.FormatInt(initialID, 10)
	req, err := http.NewRequest(http.MethodDelete, reqPath, nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify item is deleted from DB
	_, err = database.GetItem(db, initialID)
	assert.Error(t, err) // Expect an error (sql.ErrNoRows)
}

// Placeholder for TestCreateItemHandler with bad request (e.g. invalid JSON)
func TestCreateItemHandler_BadRequest(t *testing.T) {
	db := setupHandlerTestDB(t)
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/items", CreateItemHandler(db)).Methods(http.MethodPost)

	// Invalid JSON payload
	req, err := http.NewRequest(http.MethodPost, "/items", strings.NewReader("{name: \"bad json\"}"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var errResp map[string]string
	err = json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "Invalid request payload")
}
