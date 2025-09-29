package main

import (
	"fmt"
	"log"
	"net/http"

	"gotemp/internal/dbconfig"
	"gotemp/internal/handlers"
	"gotemp/internal/routes"
	"gotemp/internal/store"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration file
	config, err := dbconfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	db := dbconfig.ConnectDB(config.DatabaseURL)
	defer db.Close()

	// Connect to redis
	rdb := dbconfig.ConnectRedis()
	defer func(rdb *redis.Client) {
		_ = rdb.Close()
	}(rdb)

	// Initialize sqlc queries.
	queries := store.New(db)

	// Create a new handler with queries
	handler := handlers.NewHandlers(db, queries, rdb)

	// Setup HTTP server and routes
	mux := http.NewServeMux()

	// Setup routes without the prefix
	routes.SetupRoutes(mux, handler)

	serverAddr := fmt.Sprintf(":%s", config.ServerPort)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	// Start server in a goroutine
	fmt.Printf(" 🚀Starting server on %s\n", serverAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
