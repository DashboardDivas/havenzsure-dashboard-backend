package server

import (
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port string
}

func NewServer() *http.Server {
	port := os.Getenv("PORT")
	NewServer := &Server{
		port: port,
	}

	server := &http.Server{
		Addr:         ":" + NewServer.port,
		Handler:      NewServer.RegisterRoutes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return server
}
