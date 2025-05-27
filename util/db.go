package util

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// InitDB initializes the database connection pool
func InitDB() error {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/radius?sslmode=disable"
		log.Printf("Using default database connection string: %s", connString)
	} else {
		log.Printf("Using database connection string from environment")
	}

	var err error
	pool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to ping database: %v", err)
	}

	log.Printf("Successfully connected to database with connection pool")
	return nil
}

// GetDB returns the database connection pool
func GetDB() *pgxpool.Pool {
	if pool == nil {
		log.Printf("Warning: Database connection pool is nil")
	}
	return pool
}

// CloseDB closes the database connection pool
func CloseDB() {
	if pool != nil {
		pool.Close()
		log.Printf("Database connection pool closed")
	}
} 