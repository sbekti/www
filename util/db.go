package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database connection defaults
const (
	DefaultDBUsername = "postgres"
	DefaultDBPassword = "postgres"
	DefaultDBHost     = "localhost"
	DefaultDBPort     = "5432"
	RadiusDBName      = "radius" // Database name for device manager tool
)

// AllDatabases contains all database names that should be initialized
var AllDatabases = []string{
	RadiusDBName,
	// Add more database names here as new tools are added
}

// DBManager manages multiple database connections
type DBManager struct {
	pools map[string]*pgxpool.Pool
	mutex sync.RWMutex
}

var dbManager = &DBManager{
	pools: make(map[string]*pgxpool.Pool),
}

// DBConfig holds database connection configuration
type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

// GetDBConfigFromEnv creates a DBConfig from environment variables with defaults
func GetDBConfigFromEnv(dbName string) *DBConfig {
	config := &DBConfig{
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   dbName,
	}

	// Set defaults if not provided
	if config.Username == "" {
		config.Username = DefaultDBUsername
	}
	if config.Password == "" {
		config.Password = DefaultDBPassword
	}
	if config.Host == "" {
		config.Host = DefaultDBHost
	}
	if config.Port == "" {
		config.Port = DefaultDBPort
	}

	return config
}

// CreateConnectionPool creates a new database connection pool with the given config
func CreateConnectionPool(config *DBConfig) (*pgxpool.Pool, error) {
	// Construct connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.DBName)

	log.Printf("Connecting to database: %s@%s:%s/%s", config.Username, config.Host, config.Port, config.DBName)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	log.Printf("Successfully connected to database %s with connection pool", config.DBName)
	return pool, nil
}

// InitAllDatabases initializes connection pools for all known databases
func InitAllDatabases() error {
	log.Printf("Initializing all databases...")
	
	for _, dbName := range AllDatabases {
		if err := InitDatabase(dbName); err != nil {
			return fmt.Errorf("failed to initialize database %s: %v", dbName, err)
		}
	}
	
	log.Printf("Successfully initialized %d databases", len(AllDatabases))
	return nil
}

// InitDatabase initializes a database connection pool for the given database name
func InitDatabase(dbName string) error {
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	// Check if already initialized
	if _, exists := dbManager.pools[dbName]; exists {
		log.Printf("Database %s already initialized", dbName)
		return nil
	}

	config := GetDBConfigFromEnv(dbName)
	pool, err := CreateConnectionPool(config)
	if err != nil {
		return fmt.Errorf("failed to initialize database %s: %v", dbName, err)
	}

	dbManager.pools[dbName] = pool
	return nil
}

// GetDatabase returns the database connection pool for the given database name
func GetDatabase(dbName string) *pgxpool.Pool {
	dbManager.mutex.RLock()
	defer dbManager.mutex.RUnlock()

	pool, exists := dbManager.pools[dbName]
	if !exists {
		log.Printf("Warning: Database %s connection pool not found", dbName)
		return nil
	}
	return pool
}

// CloseDatabase closes the database connection pool for the given database name
func CloseDatabase(dbName string) {
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	if pool, exists := dbManager.pools[dbName]; exists {
		pool.Close()
		delete(dbManager.pools, dbName)
		log.Printf("Database %s connection pool closed", dbName)
	}
}

// CloseAllDatabases closes all database connection pools
func CloseAllDatabases() {
	dbManager.mutex.Lock()
	defer dbManager.mutex.Unlock()

	for dbName, pool := range dbManager.pools {
		pool.Close()
		log.Printf("Database %s connection pool closed", dbName)
	}
	dbManager.pools = make(map[string]*pgxpool.Pool)
} 