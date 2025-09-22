package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents the database service interact with PostgreSQL
type Service interface {

	// Close the database connection
	CloseDB()
}

type service struct {
	pool *pgxpool.Pool
}

var (
	host      = os.Getenv("DB_HOST")
	port      = os.Getenv("DB_PORT")
	user      = os.Getenv("DB_APP_USER")
	password  = os.Getenv("DB_APP_PASSWORD")
	database  = os.Getenv("DB_NAME")
	schema    = os.Getenv("DB_SCHEMA")
	dbService *service
)

func InitDB() (Service, error) {

	//Reuse existing connection if already initialized
	if dbService != nil {
		return dbService, nil
	}

	// Create the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		host, port, user, password, database, schema)

	// Open connections to the database
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	// Set context with timeout for pinging the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Ping the database to check connectivity
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	dbService = &service{pool: pool}
	return dbService, nil
}

func (s *service) CloseDB() {
	log.Printf("Disconnected from database: %s", database)
	s.pool.Close()
}
