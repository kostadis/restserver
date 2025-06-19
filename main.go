package main

import (
	"database/sql"
	"log"
	"net/http"
	"time" // Added for loggingMiddleware

	"app/database" // Assuming module name 'app'
	"app/handlers" // Assuming module name 'app'

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // Ensure SQLite driver is included
)

var DB *sql.DB

func init() {
	var err error
	DB, err = database.InitDB("sqlite.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// loggingMiddleware logs the incoming request and the time taken to process it.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.RequestURI)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Register API routes
	// Item routes
	router.HandleFunc("/items", handlers.CreateItemHandler(DB)).Methods(http.MethodPost)
	router.HandleFunc("/items/{id}", handlers.GetItemHandler(DB)).Methods(http.MethodGet)
	// Note: For GET /items and GET /items/{id} to work correctly with gorilla/mux,
	// ensure the more specific route (/items/{id}) is registered before the general one (/items)
	// if there's any ambiguity, or rely on mux's default matching.
	// However, with distinct paths or methods, order is less critical.
	// The current setup for /items (GET all) and /items/{id} (GET one) is fine as they are distinct.
	// If GetItemsHandler was meant for a path like "/items/" and GetItemHandler for "/items/{id}",
	// then specific registration order or strict slash handling might be needed.
	// Given they are:
	// POST /items -> CreateItemHandler
	// GET  /items/{id} -> GetItemHandler
	// GET  /items -> GetItemsHandler
	// PUT  /items/{id} -> UpdateItemHandler
	// DELETE /items/{id} -> DeleteItemHandler
	// This is a standard and correct setup.
	router.HandleFunc("/items", handlers.GetItemsHandler(DB)).Methods(http.MethodGet)
	router.HandleFunc("/items/{id}", handlers.UpdateItemHandler(DB)).Methods(http.MethodPut)
	router.HandleFunc("/items/{id}", handlers.DeleteItemHandler(DB)).Methods(http.MethodDelete)

	// Apply the logging middleware to the main router
	loggedRouter := loggingMiddleware(router)

	// Start the HTTP server
	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	// Use the middleware-wrapped router for ListenAndServe
	if err := http.ListenAndServe(port, loggedRouter); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
