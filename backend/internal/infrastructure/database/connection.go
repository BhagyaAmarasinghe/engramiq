package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/engramiq/engramiq-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New creates a new database connection with proper configuration
// We're using GORM as our ORM for better developer experience while maintaining performance
func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger based on environment
	logConfig := logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}

	// Enable info logging in development
	if cfg.URL == "" || cfg.URL == "development" {
		logConfig.LogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // Use standard log writer
			logConfig,
		),
		PrepareStmt:            true, // Prepare statements for better performance
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL database: %w", err)
	}

	// Configure connection pool for optimal performance
	// These settings prevent connection exhaustion and improve response times
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxConnections / 2)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable pgvector extension for semantic search
	// This is crucial for our AI-powered search features
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return nil, fmt.Errorf("failed to create pgvector extension: %w", err)
	}

	// Enable UUID generation
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid extension: %w", err)
	}

	return db, nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	return sqlDB.PingContext(ctx)
}