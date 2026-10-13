package database

import (
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitPostgres initializes the PostgreSQL database connection
func InitPostgres() {
	var err error

	// Get database connection parameters from environment variables
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "talehaven")
	password := getEnv("DB_PASSWORD", "talehaven123")
	dbname := getEnv("DB_NAME", "talehaven")
	sslmode := getEnv("DB_SSLMODE", "disable")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Open database connection
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		utils.Logger.Fatal("Failed to open database connection", zap.Error(err))
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)                 // Maximum number of open connections
	DB.SetMaxIdleConns(5)                  // Maximum number of idle connections
	DB.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	// Test database connection
	err = DB.Ping()
	if err != nil {
		utils.Logger.Fatal("Failed to ping database", zap.Error(err))
	}

	utils.Logger.Info("PostgreSQL connection established successfully",
		zap.String("host", host),
		zap.String("port", port),
		zap.String("database", dbname))

	// Run database migrations
	RunMigrations()
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			utils.Logger.Error("Error closing database connection", zap.Error(err))
		} else {
			utils.Logger.Info("Database connection closed")
		}
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

