package database

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID        string
	Name      string
	Up        func(*gorm.DB) error
	Down      func(*gorm.DB) error
	Timestamp time.Time
}

// MigrationRecord tracks applied migrations in the database
type MigrationRecord struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *gorm.DB) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// RegisterMigration adds a migration to the runner
func (mr *MigrationRunner) RegisterMigration(migration Migration) {
	mr.migrations = append(mr.migrations, migration)
}

// RunMigrations executes all pending migrations
func (mr *MigrationRunner) RunMigrations() error {
	// Ensure migration table exists
	if err := mr.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Sort migrations by timestamp
	sort.Slice(mr.migrations, func(i, j int) bool {
		return mr.migrations[i].Timestamp.Before(mr.migrations[j].Timestamp)
	})

	// Get applied migrations
	var appliedMigrations []MigrationRecord
	if err := mr.db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create map of applied migration IDs for quick lookup
	applied := make(map[string]bool)
	for _, record := range appliedMigrations {
		applied[record.ID] = true
	}

	// Run pending migrations
	for _, migration := range mr.migrations {
		if applied[migration.ID] {
			continue // Skip already applied migration
		}

		fmt.Printf("Running migration: %s - %s\n", migration.ID, migration.Name)
		
		// Begin transaction
		tx := mr.db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}

		// Run migration
		if err := migration.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to run migration %s: %w", migration.ID, err)
		}

		// Record migration
		record := MigrationRecord{
			ID:        migration.ID,
			Name:      migration.Name,
			AppliedAt: time.Now(),
		}
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.ID, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.ID, err)
		}

		fmt.Printf("✅ Completed migration: %s\n", migration.ID)
	}

	return nil
}

// RollbackMigration rolls back a specific migration
func (mr *MigrationRunner) RollbackMigration(migrationID string) error {
	// Find the migration
	var migration Migration
	found := false
	for _, m := range mr.migrations {
		if m.ID == migrationID {
			migration = m
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("migration %s not found", migrationID)
	}

	// Check if migration was applied
	var record MigrationRecord
	if err := mr.db.Where("id = ?", migrationID).First(&record).Error; err != nil {
		return fmt.Errorf("migration %s was not applied", migrationID)
	}

	fmt.Printf("Rolling back migration: %s - %s\n", migration.ID, migration.Name)

	// Begin transaction
	tx := mr.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Run rollback
	if err := migration.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration %s: %w", migration.ID, err)
	}

	// Remove migration record
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record %s: %w", migration.ID, err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", migration.ID, err)
	}

	fmt.Printf("✅ Rolled back migration: %s\n", migration.ID)
	return nil
}

// GetAppliedMigrations returns list of applied migrations
func (mr *MigrationRunner) GetAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	if err := mr.db.Order("applied_at").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetPendingMigrations returns list of pending migrations
func (mr *MigrationRunner) GetPendingMigrations() ([]Migration, error) {
	// Get applied migrations
	appliedMigrations, err := mr.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	// Create map of applied migration IDs
	applied := make(map[string]bool)
	for _, record := range appliedMigrations {
		applied[record.ID] = true
	}

	// Filter pending migrations
	var pending []Migration
	for _, migration := range mr.migrations {
		if !applied[migration.ID] {
			pending = append(pending, migration)
		}
	}

	// Sort by timestamp
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Timestamp.Before(pending[j].Timestamp)
	})

	return pending, nil
}