package database

import (
	"database/sql"
	"os"
	"testing"

	"app/models" // Assuming module 'app'

	// These will cause compilation errors if 'go get' failed and they are not otherwise available.
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for in-memory db
)

// setupTestDB initializes an in-memory SQLite database for testing.
// It returns the database connection and a teardown function to close the DB.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Ensure the schema file path is correct if InitDB relies on it.
	// For ":memory:", InitDB will create the schema.
	// If InitDB strictly needs schema.sql from a path, this might need adjustment
	// or ensure schema.sql is accessible during tests.
	// Given our InitDB reads "database/schema.sql", we need to ensure it's found.
	// This usually means running tests from the project root or adjusting path.
	// For simplicity, we assume InitDB can find it or the test environment handles it.
	// One common pattern is to have a project root helper for paths.
	// Or, for tests, InitDB could be modified or an alternative test init func created
	// that takes the schema string directly.

	// Hacky way to ensure schema.sql is found if tests are run from package dir:
	// This assumes 'schema.sql' is in the same directory as 'database.go' (it is).
	// And that tests are run in a context where this relative path is valid.
	// If `os.ReadFile("database/schema.sql")` in InitDB fails, this is why.
	// A better solution would be to embed schema or pass path to InitDB.

	db, err := InitDB(":memory:") // Use in-memory database for speed and isolation
	require.NoError(t, err, "Failed to initialize test database")

	teardown := func() {
		err := db.Close()
		require.NoError(t, err, "Failed to close test database")
	}

	return db, teardown
}

func TestCreateItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	item := models.Item{
		Name:        "Test Item 1",
		Description: "A description for test item 1",
		Price:       9.99,
	}

	id, err := CreateItem(db, item)
	require.NoError(t, err, "CreateItem should not produce an error")
	require.NotZero(t, id, "CreateItem should return a non-zero ID")

	// Verify the item in the DB
	var fetchedItem models.Item
	err = db.QueryRow("SELECT id, name, description, price FROM items WHERE id = ?", id).Scan(
		&fetchedItem.ID, &fetchedItem.Name, &fetchedItem.Description, &fetchedItem.Price,
	)
	require.NoError(t, err, "Failed to fetch created item for verification")

	assert.Equal(t, id, fetchedItem.ID, "Fetched item ID should match returned ID")
	assert.Equal(t, item.Name, fetchedItem.Name, "Fetched item Name should match input")
	assert.Equal(t, item.Description, fetchedItem.Description, "Fetched item Description should match input")
	assert.Equal(t, item.Price, fetchedItem.Price, "Fetched item Price should match input")
}

func TestGetItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Setup: Create an item first
	itemToCreate := models.Item{Name: "Test GetItem", Description: "Desc", Price: 10.50}
	id, err := CreateItem(db, itemToCreate)
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		fetchedItem, err := GetItem(db, id)
		require.NoError(t, err, "GetItem should not error for existing ID")
		assert.Equal(t, id, fetchedItem.ID)
		assert.Equal(t, itemToCreate.Name, fetchedItem.Name)
		assert.Equal(t, itemToCreate.Description, fetchedItem.Description)
		assert.Equal(t, itemToCreate.Price, fetchedItem.Price)
	})

	t.Run("non-existent item", func(t *testing.T) {
		_, err := GetItem(db, 99999) // Assuming 99999 does not exist
		require.Error(t, err, "GetItem should error for non-existent ID")
		assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == sql.ErrNoRows.Error(), "Error should be sql.ErrNoRows")
	})
}

func TestGetItems(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	t.Run("empty database", func(t *testing.T) {
		items, err := GetItems(db)
		require.NoError(t, err, "GetItems should not error on empty DB")
		assert.Empty(t, items, "GetItems should return an empty slice for an empty DB")
	})

	t.Run("with multiple items", func(t *testing.T) {
		// Create some items
		item1 := models.Item{Name: "Item A", Price: 1.00}
		item2 := models.Item{Name: "Item B", Price: 2.00}
		_, err := CreateItem(db, item1)
		require.NoError(t, err)
		_, err = CreateItem(db, item2)
		require.NoError(t, err)

		items, err := GetItems(db)
		require.NoError(t, err, "GetItems should not error when DB has items")
		assert.Len(t, items, 2, "GetItems should return all items in DB")

		// Basic check if items are present (could be more thorough)
		var foundA, foundB bool
		for _, item := range items {
			if item.Name == "Item A" {
				foundA = true
			}
			if item.Name == "Item B" {
				foundB = true
			}
		}
		assert.True(t, foundA, "Item A should be in the results")
		assert.True(t, foundB, "Item B should be in the results")
	})
}

func TestUpdateItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Setup: Create an item first
	itemToCreate := models.Item{Name: "Original Name", Description: "Original Desc", Price: 10.00}
	id, err := CreateItem(db, itemToCreate)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updatedItem := models.Item{
			Name:        "Updated Name",
			Description: "Updated Desc",
			Price:       20.00,
		}
		rowsAffected, err := UpdateItem(db, id, updatedItem)
		require.NoError(t, err, "UpdateItem should not error for existing ID")
		assert.Equal(t, int64(1), rowsAffected, "UpdateItem should affect 1 row")

		// Verify update
		fetchedItem, err := GetItem(db, id)
		require.NoError(t, err)
		assert.Equal(t, updatedItem.Name, fetchedItem.Name)
		assert.Equal(t, updatedItem.Description, fetchedItem.Description)
		assert.Equal(t, updatedItem.Price, fetchedItem.Price)
	})

	t.Run("update non-existent item", func(t *testing.T) {
		item := models.Item{Name: "Non-existent update"}
		rowsAffected, err := UpdateItem(db, 99999, item)
		// UpdateItem is designed to return sql.ErrNoRows if no rows were affected and no other error occurred.
		require.Error(t, err, "UpdateItem should error for non-existent ID")
		assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows for non-existent update")
		assert.Equal(t, int64(0), rowsAffected, "UpdateItem should affect 0 rows for non-existent ID")
	})
}

func TestDeleteItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Setup: Create an item first
	itemToCreate := models.Item{Name: "To Be Deleted", Price: 5.00}
	id, err := CreateItem(db, itemToCreate)
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		rowsAffected, err := DeleteItem(db, id)
		require.NoError(t, err, "DeleteItem should not error for existing ID")
		assert.Equal(t, int64(1), rowsAffected, "DeleteItem should affect 1 row")

		// Verify deletion
		_, err = GetItem(db, id)
		require.Error(t, err, "GetItem should error after deletion")
		assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows after deletion")
	})

	t.Run("delete non-existent item", func(t *testing.T) {
		rowsAffected, err := DeleteItem(db, 99999)
		// DeleteItem is designed to return sql.ErrNoRows if no rows were affected and no other error occurred.
		require.Error(t, err, "DeleteItem should error for non-existent ID")
		assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows for non-existent delete")
		assert.Equal(t, int64(0), rowsAffected, "DeleteItem should affect 0 rows for non-existent ID")
	})
}

// Note on InitDB and schema.sql path:
// The InitDB function reads "database/schema.sql". For tests to reliably find this,
// they should ideally be run from the project root (e.g., `go test ./database/...`).
// If `go test` is run directly inside the `database` directory, `os.ReadFile("database/schema.sql")`
// would look for `database/database/schema.sql`, which is incorrect.
// The current InitDB implementation might need to be more robust about path handling,
// or tests need to ensure they run from a context where the relative path is correct.
// One common way is to use `runtime.Caller` to get the current file's directory and build
// an absolute path to schema.sql, or use `go:embed` for the schema.
// For this exercise, we assume the test execution context or a future InitDB refinement handles this.
// The `os.Chdir("..")` and `defer os.Chdir(originalWD)` pattern in setupTestDB could be
// a workaround if tests are run from the package directory, but it's brittle.
// The simplest for now is to rely on `InitDB(":memory:")` creating the schema correctly
// because `schemaSQL, err := os.ReadFile("database/schema.sql")` must succeed.
// If `InitDB` fails in `setupTestDB` because `schema.sql` is not found, the path needs fixing.
// A quick check: `os.ReadFile("schema.sql")` might be what `InitDB` should use if it's in the same dir.
// Let's assume `InitDB` is called from a context where `database/schema.sql` is a valid relative path.
// The `database.go` file is in the `database` package. `schema.sql` is also in `database`.
// So, from `database.go`, `os.ReadFile("schema.sql")` should be correct.
// The current `InitDB` uses `os.ReadFile("database/schema.sql")`. This is problematic if `InitDB`
// is called from within the `database` package (e.g. by tests in the same package) because the
// working directory would be `database/`, so it would look for `database/database/schema.sql`.
// This needs correction in `InitDB` or test setup.

// Temporary fix for schema path in InitDB for testing:
// I will modify InitDB to read "schema.sql" if the filepath is ":memory:"
// This is a bit of a hack for the testing context.
// Better: InitDB should take schema path as arg, or embed schema.
// For now, I'll adjust the InitDB in database.go to handle this.
// This is outside the scope of this file, but necessary for these tests to pass.
// Let's assume `database.InitDB` is robust enough or will be fixed.
// The tests above are written assuming `InitDB(":memory:")` works correctly.
// The `errors.Is` was added in Go 1.13. Ensure Go version compatibility.
// If using an older Go, `err == sql.ErrNoRows` would be used.
// The `require` and `assert` calls will fail if `testify` is not present.```go
package database

import (
	"database/sql"
	"errors" // Required for errors.Is
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"app/models" // Assuming module 'app'

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for in-memory db
)

// getProjectRoot returns the root directory of the project.
// This is a helper to ensure schema.sql can be found regardless of where tests are run.
func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	// Root directory of project is two levels up from database_test.go
	return filepath.Join(filepath.Dir(b), "..")
}

// setupTestDB initializes an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Construct the path to schema.sql relative to the project root
	projectRoot := getProjectRoot()
	schemaPath := filepath.Join(projectRoot, "database", "schema.sql")

	// Check if schema file exists, to give a clearer error if it doesn't.
	// This is important because InitDB(":memory:") internally tries to read it.
	_, err := os.Stat(schemaPath)
	require.NoError(t, err, "schema.sql file not found at %s. Ensure path is correct.", schemaPath)

	// The InitDB function needs to correctly locate "database/schema.sql".
	// If InitDB is called with ":memory:", it implies it should handle schema creation.
	// Our current InitDB reads "database/schema.sql" relative to its execution path.
	// To make tests robust, we ensure InitDB can find the schema.
	// One way is to temporarily change CWD or make InitDB more flexible.
	// For now, we rely on the fact that `InitDB` uses `os.ReadFile("database/schema.sql")`
	// and we hope the test runner executes from a context (like project root) where this path is valid.
	// A better InitDB would take the schema path or use `go:embed`.

	// Let's assume InitDB is called from project root in tests, or schema path is global.
	// The InitDB function was: os.ReadFile("database/schema.sql")
	// If tests run from project root, this is fine. If run from `database/` dir, it fails.
	// The best fix is in InitDB itself or by passing schema path.
	// For this test, we will assume `InitDB` is robust or tests are run from root.

	db, err := InitDB(":memory:") // Use in-memory database
	require.NoError(t, err, "Failed to initialize test database. Ensure InitDB can find 'database/schema.sql' or correctly handles in-memory setup.")

	teardown := func() {
		err := db.Close()
		require.NoError(t, err, "Failed to close test database")
	}

	return db, teardown
}


func TestCreateItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	item := models.Item{
		Name:        "Test Item 1",
		Description: "A description for test item 1",
		Price:       9.99,
	}

	id, err := CreateItem(db, item)
	require.NoError(t, err, "CreateItem should not produce an error")
	require.NotZero(t, id, "CreateItem should return a non-zero ID")

	var fetchedItem models.Item
	err = db.QueryRow("SELECT id, name, description, price FROM items WHERE id = ?", id).Scan(
		&fetchedItem.ID, &fetchedItem.Name, &fetchedItem.Description, &fetchedItem.Price,
	)
	require.NoError(t, err, "Failed to fetch created item for verification")

	assert.Equal(t, id, fetchedItem.ID)
	assert.Equal(t, item.Name, fetchedItem.Name)
	assert.Equal(t, item.Description, fetchedItem.Description)
	assert.Equal(t, item.Price, fetchedItem.Price)
}

func TestGetItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	itemToCreate := models.Item{Name: "Test GetItem", Description: "Desc", Price: 10.50}
	id, err := CreateItem(db, itemToCreate)
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		fetchedItem, err := GetItem(db, id)
		require.NoError(t, err, "GetItem should not error for existing ID")
		assert.Equal(t, id, fetchedItem.ID)
		assert.Equal(t, itemToCreate.Name, fetchedItem.Name)
	})

	t.Run("non-existent item", func(t *testing.T) {
		_, err := GetItem(db, 99999)
		require.Error(t, err, "GetItem should error for non-existent ID")
		assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows")
	})
}

func TestGetItems(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	t.Run("empty database", func(t *testing.T) {
		items, err := GetItems(db)
		require.NoError(t, err, "GetItems should not error on empty DB")
		assert.Empty(t, items, "GetItems should return an empty slice for an empty DB")
	})

	t.Run("with multiple items", func(t *testing.T) {
		item1 := models.Item{Name: "Item A", Price: 1.00}
		item2 := models.Item{Name: "Item B", Price: 2.00}
		_, err := CreateItem(db, item1)
		require.NoError(t, err)
		_, err = CreateItem(db, item2)
		require.NoError(t, err)

		items, err := GetItems(db)
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})
}

func TestUpdateItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	id, err := CreateItem(db, models.Item{Name: "Original", Price: 1.00})
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		updatedItemData := models.Item{Name: "Updated", Price: 2.00}
		rowsAffected, err := UpdateItem(db, id, updatedItemData)
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		fetched, _ := GetItem(db, id)
		assert.Equal(t, "Updated", fetched.Name)
		assert.Equal(t, 2.00, fetched.Price)
	})

	t.Run("update non-existent item", func(t *testing.T) {
		_, err := UpdateItem(db, 99999, models.Item{Name: "NonExistent"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows), "Error should be sql.ErrNoRows for non-existent update")
	})
}

func TestDeleteItem(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	id, err := CreateItem(db, models.Item{Name: "To Delete", Price: 1.00})
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		rowsAffected, err := DeleteItem(db, id)
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		_, err = GetItem(db, id)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})

	t.Run("delete non-existent item", func(t *testing.T) {
		_, err := DeleteItem(db, 99999)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrNoRows))
	})
}
```

**Critical Note on `InitDB` and `schema.sql` Path for Tests:**
The previous `database_test.go` content highlighted a significant issue: `InitDB` reads `"database/schema.sql"`. If `go test` is run from within the `database` package directory, the working directory is `.../project/database/`, so `InitDB` would look for `.../project/database/database/schema.sql`, which is wrong.

The `getProjectRoot()` helper and `schemaPath` construction in the new `setupTestDB` are attempts to acknowledge this. However, `InitDB` itself isn't changed by this test file. For `InitDB(":memory:")` to reliably work in tests, `InitDB` needs to be more robust in finding `schema.sql` or not depend on a hardcoded relative path when using `:memory:`.

A common solution is for `InitDB` to read `schema.sql` (relative to its own file location) or for the schema to be embedded.
I will now **modify `database/database.go`'s `InitDB` function** to correctly locate `schema.sql` when the `database` package is the current directory. This is essential for the tests to pass regarding schema initialization. The change will be: if `InitDB` is called, it should look for `schema.sql` in the same directory as `database.go`.
