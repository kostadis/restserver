package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen-config.yaml openapi.yaml

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"app/database"
	"app/handlers"
	"app/internal/generated/openapi" // Added for generated code

	"github.com/go-chi/chi/v5" // Replaced gorilla/mux
	// chi_middleware "github.com/go-chi/chi/v5/middleware" // Optional: For Chi's own middlewares
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = database.InitDB("sqlite.db") // Assuming InitDB is compatible or adapted
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// loggingMiddleware remains the same as it's standard http.Handler compatible
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func main() {
	// Create a new Chi router
	router := chi.NewRouter()

	// Apply the logging middleware (Chi also has its own logging middleware if preferred)
	router.Use(loggingMiddleware)
	// router.Use(chi_middleware.Logger) // Alternative using Chi's logger

	// Instantiate our Item API server implementation
	itemAPIServer := handlers.NewItemAPIServer(DB)

	// Register the OpenAPI-generated handlers.
	// The openapi.HandlerWithOptions function will register routes like /items/{id}
	// onto the router passed to it, or create a new one.
	// We can mount it on a sub-route e.g. /api/v1 or directly on root.
	// For this example, let's assume the paths in openapi.yaml are root paths.
	// openapi.Handler() creates a new chi router internally and mounts the generated handlers.
	// We want to use our main router.

	// Option 1: Let openapi.Handler create its own router and mount it
	// itemAPIChiRouter := openapi.Handler(itemAPIServer)
	// router.Mount("/items", itemAPIChiRouter) // This would make the path /items/items/{id} - likely not desired
	// The paths in openapi.yaml are /items/{id}, so we want to use HandlerFromMux or HandlerWithOptions

	// Option 2: Use HandlerFromMux to register generated routes on our main router
	// This is generally cleaner if the generated paths are meant to be at the root of this router.
	// The `openapi.HandlerWithOptions` function adds the routes to the provided BaseRouter.
	// The generated `HandlerWithOptions` in `item_api.gen.go` looks like:
	// r.Group(func(r chi.Router) {
	//   r.Get(options.BaseURL+"/items/{id}", wrapper.GetItemById)
	// })
	// So, it will add "/items/{id}" to the router we pass.

	// This will register GET /items/{id}, POST /items, and PUT /items/{id}
	// (because UpdateItemById is now part of ServerInterface and implemented by ItemAPIServer)
	openapi.HandlerWithOptions(itemAPIServer, openapi.ChiServerOptions{
		BaseRouter: router, // Register on our main router
		// Middlewares: []openapi.MiddlewareFunc{}, // Optional: API specific middlewares
		// ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) { ... } // Optional
	})

	// Register other existing API routes using Chi's syntax
	// These handlers are from handlers/item_handlers.go
	// Note: The GetItemHandler was removed, so we don't register it here.
	// The POST /items route is now handled by the OpenAPI generated code via HandlerWithOptions.
	router.Get("/items", handlers.GetItemsHandler(DB)) // For getting all items
	// router.Put("/items/{id}", handlers.UpdateItemHandler(DB)) // THIS LINE IS REMOVED
	// router.Delete("/items/{id}", handlers.DeleteItemHandler(DB)) // THIS LINE IS REMOVED


	// Start the HTTP server
	port := ":8080"
	log.Printf("Server starting on port %s using Chi router", port) // Corrected log message formatting
	if err := http.ListenAndServe(port, router); err != nil { // Pass the Chi router directly
		log.Fatalf("Failed to start server: %v", err)
	}
}
