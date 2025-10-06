package server

import (
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port string
}

func NewServer(db *pgxpool.Pool) *http.Server {
	port := os.Getenv("PORT")
	NewServer := &Server{
		port: port,
	}

	server := &http.Server{
		Addr:         ":" + NewServer.port,
		Handler:      NewServer.RegisterRoutes(db),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return server
}
