// citaiontion: The code inside main.go/db.go/server.go/router.go are adapted from
// https://github.com/Melkeydev/go-blueprint (go-blueprint library)

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/database"
	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/platform/server"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	// Initialize the database connection
	databaseService, dbError := database.InitDB()
	if dbError != nil {
		log.Fatalf("Unable to connect to database: %v\n", dbError)
	}
	defer databaseService.CloseDB()

	// Initialize the HTTP server
	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	// Start the HTTP server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")

}

func gracefulShutdown(srv *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// Ensure the stop function is called to release resources
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true

}
