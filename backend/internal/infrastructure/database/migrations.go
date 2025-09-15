package database

import (
	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/engramiq/engramiq-backend/internal/infrastructure/database/migrations"
	"gorm.io/gorm"
)

// Migrate runs all database migrations
// This includes both schema migrations (GORM auto-migrate) and data migrations
func Migrate(db *gorm.DB) error {
	// Create custom types first
	if err := createCustomTypes(db); err != nil {
		return err
	}

	// Auto-migrate all domain models
	// Order matters here due to foreign key constraints
	models := []interface{}{
		// Core models first
		&domain.User{},
		&domain.RefreshToken{},
		&domain.Site{},
		
		// Component models
		&domain.SiteComponent{},
		&domain.ComponentRelationship{},
		
		// Document and processing models
		&domain.Document{},
		&domain.ExtractedAction{},
		&domain.ActionComponent{},
		
		// Event and timeline models
		&domain.SiteEvent{},
		
		// Query models
		&domain.UserQuery{},
		&domain.QuerySource{},
		
		// Analytics models
		&domain.QueryAnalytics{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return err
	}

	// Create triggers for computed columns
	if err := createTriggers(db); err != nil {
		return err
	}

	// Run data migrations
	runner := NewMigrationRunner(db)
	migrationsList := migrations.GetAllMigrations()
	
	// Convert migrations to database.Migration and register them
	for _, m := range migrationsList {
		dbMigration := Migration{
			ID:        m.ID,
			Name:      m.Name,
			Up:        m.Up,
			Down:      m.Down,
			Timestamp: m.Timestamp,
		}
		runner.RegisterMigration(dbMigration)
	}
	
	if err := runner.RunMigrations(); err != nil {
		return err
	}

	return nil
}

// createCustomTypes creates PostgreSQL ENUM types
func createCustomTypes(db *gorm.DB) error {
	// Component type enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE component_type AS ENUM (
			'inverter', 'combiner', 'panel', 'transformer', 
			'meter', 'switchgear', 'monitoring', 'other'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Document type enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE document_type AS ENUM (
			'field_service_report', 'email', 'meeting_transcript',
			'work_order', 'inspection_report', 'warranty_claim',
			'contract', 'manual', 'drawing', 'other'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Action type enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE action_type AS ENUM (
			'maintenance', 'replacement', 'troubleshoot', 'inspection',
			'repair', 'testing', 'installation', 'commissioning',
			'fault_clearing', 'monitoring', 'cleaning', 'other'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Action status enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE action_status AS ENUM (
			'planned', 'in_progress', 'completed', 'cancelled',
			'on_hold', 'requires_follow_up'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Event type enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE event_type AS ENUM (
			'maintenance_scheduled', 'maintenance_completed',
			'fault_occurred', 'fault_cleared',
			'replacement_scheduled', 'replacement_completed',
			'inspection_scheduled', 'inspection_completed',
			'warranty_claim', 'performance_alert',
			'contract_milestone', 'other'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Event priority enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE event_priority AS ENUM ('low', 'medium', 'high', 'critical');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	// Relationship type enum
	db.Exec(`DO $$ BEGIN
		CREATE TYPE relationship_type AS ENUM (
			'connects_to', 'powers', 'controls', 'monitors',
			'parent_child', 'same_string', 'same_combiner'
		);
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;`)

	return nil
}

// createIndexes creates additional indexes for performance
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// Full-text search indexes
		`CREATE INDEX IF NOT EXISTS idx_documents_fts ON documents 
		 USING gin(to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(processed_content, '')))`,
		
		// Vector similarity search indexes (requires pgvector)
		`CREATE INDEX IF NOT EXISTS idx_documents_embedding ON documents 
		 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`,
		
		`CREATE INDEX IF NOT EXISTS idx_components_embedding ON site_components 
		 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`,
		
		`CREATE INDEX IF NOT EXISTS idx_actions_embedding ON extracted_actions 
		 USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`,
		
		// JSONB indexes for metadata queries
		`CREATE INDEX IF NOT EXISTS idx_components_specifications ON site_components USING gin(specifications)`,
		`CREATE INDEX IF NOT EXISTS idx_components_electrical_data ON site_components USING gin(electrical_data)`,
		`CREATE INDEX IF NOT EXISTS idx_actions_measurements ON extracted_actions USING gin(measurements)`,
		
		// Composite indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_components_site_type ON site_components(site_id, component_type)`,
		`CREATE INDEX IF NOT EXISTS idx_events_site_timeline ON site_events(site_id, start_time, end_time)`,
		`CREATE INDEX IF NOT EXISTS idx_actions_site_date ON extracted_actions(site_id, action_date)`,
		
		// Array indexes
		`CREATE INDEX IF NOT EXISTS idx_actions_technicians ON extracted_actions USING gin(technician_names)`,
		`CREATE INDEX IF NOT EXISTS idx_events_affected_components ON site_events USING gin(affected_component_ids)`,
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			// Log but don't fail - some indexes might already exist
			continue
		}
	}

	return nil
}

// createTriggers creates database triggers for computed columns
func createTriggers(db *gorm.DB) error {
	// Create function to update content_vector
	db.Exec(`
		CREATE OR REPLACE FUNCTION update_content_vector() RETURNS trigger AS $$
		BEGIN
			NEW.content_vector := to_tsvector('english', COALESCE(NEW.title, '') || ' ' || COALESCE(NEW.processed_content, ''));
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)

	// Create trigger for insert and update (separate commands)
	db.Exec(`DROP TRIGGER IF EXISTS documents_content_vector_trigger ON documents;`)
	db.Exec(`
		CREATE TRIGGER documents_content_vector_trigger
		BEFORE INSERT OR UPDATE OF title, processed_content ON documents
		FOR EACH ROW
		EXECUTE FUNCTION update_content_vector();
	`)

	// Update existing documents
	db.Exec(`
		UPDATE documents 
		SET content_vector = to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(processed_content, ''))
		WHERE content_vector IS NULL;
	`)

	return nil
}