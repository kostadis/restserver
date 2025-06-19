package database

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath" // Added for robust schema path
	"runtime"       // Added for robust schema path

	"app/models"

	_ "github.com/mattn/go-sqlite3"
)

// getPackageDir returns the directory of the current Go package.
func getPackageDir() string {
	_, b, _, _ := runtime.Caller(0) // Get information about the caller (this file)
	return filepath.Dir(b)          // Directory of this file (database package)
}

func InitDB(filepathArg string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepathArg)
	if err != nil {
		return nil, err
	}

	// Construct path to schema.sql relative to this file's location (database package directory)
	// This makes it robust to where the application or tests are run from.
	schemaPath := filepath.Join(getPackageDir(), "schema.sql")

	schemaSQL, err := os.ReadFile(schemaPath)
	if err != nil {
		db.Close()
		return nil, errors.New("failed to read schema.sql at " + schemaPath + ": " + err.Error())
	}

	_, err = db.Exec(string(schemaSQL))
	if err != nil {
		db.Close()
		return nil, errors.New("failed to execute schema: " + err.Error())
	}

	return db, nil
}

// CreateItem, GetItem, GetItems, UpdateItem, DeleteItem functions remain unchanged...
// (Assuming they are already present from previous steps)

// CreateItem adds a new item to the database.
// It returns the ID of the newly created item.
func CreateItem(db *sql.DB, item models.Item) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO items(name, description, priority) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(item.Name, item.Description, item.Priority)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetItem retrieves a single item from the database by its ID.
func GetItem(db *sql.DB, id int64) (models.Item, error) {
	stmt, err := db.Prepare("SELECT id, name, description, priority FROM items WHERE id = ?")
	if err != nil {
		return models.Item{}, err
	}
	defer stmt.Close()

	var item models.Item
	err = stmt.QueryRow(id).Scan(&item.ID, &item.Name, &item.Description, &item.Priority)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Item{}, err
		}
		return models.Item{}, err
	}
	return item, nil
}

// GetItems retrieves all items from the database.
func GetItems(db *sql.DB) ([]models.Item, error) {
	stmt, err := db.Prepare("SELECT id, name, description, priority FROM items")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Priority); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateItem modifies an existing item in the database.
func UpdateItem(db *sql.DB, id int64, item models.Item) (int64, error) {
	stmt, err := db.Prepare("UPDATE items SET name = ?, description = ?, priority = ? WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(item.Name, item.Description, item.Priority, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
    if rowsAffected == 0 {
        return 0, sql.ErrNoRows
    }
	return rowsAffected, nil
}

// DeleteItem removes an item from the database by its ID.
func DeleteItem(db *sql.DB, id int64) (int64, error) {
	stmt, err := db.Prepare("DELETE FROM items WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
    if rowsAffected == 0 {
        return 0, sql.ErrNoRows
    }
	return rowsAffected, nil
}
