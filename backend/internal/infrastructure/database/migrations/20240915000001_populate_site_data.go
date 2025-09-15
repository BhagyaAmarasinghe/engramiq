package migrations

import (
	"fmt"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
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

// CreatePopulateSiteDataMigration creates the migration for populating site S2367 with inverters
func CreatePopulateSiteDataMigration() Migration {
	return Migration{
		ID:        "20240915000001",
		Name:      "Populate site S2367 with inverter data",
		Timestamp: time.Date(2024, 9, 15, 0, 0, 1, 0, time.UTC),
		Up:       populateSiteDataUp,
		Down:     populateSiteDataDown,
	}
}

func populateSiteDataUp(tx *gorm.DB) error {
	siteID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	
	// Check if site already exists
	var existingSite domain.Site
	if err := tx.Where("site_code = ?", "S2367").First(&existingSite).Error; err == nil {
		// Site already exists, update the site ID for component creation
		siteID = existingSite.ID
	} else {
		// Create the site S2367
		installDate := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)
		site := &domain.Site{
			ID:                 siteID,
			SiteCode:           "S2367",
			Name:               "Solar Installation S2367",
			Address:            "Industrial Solar Park, Renewable Energy District",
			Country:            "US",
			TotalCapacityKW:    2850.0, // Calculated from all inverters
			NumberOfInverters:  36,
			InstallationDate:   &installDate,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := tx.Create(site).Error; err != nil {
			return err
		}
	}

	// Define inverter data from inverter_nodes.json
	inverters := []struct {
		Name         string
		Manufacturer string
		Model        string
		CapacityKW   float64
		SpatialID    string
		Location     string
	}{
		// Ground mount inverters (1-18)
		{"1", "SOLECTRIA", "PVI 75TL", 75.0, "8ed90bd7-bb30-4bd9-94fb-d051b9baccab", "Ground"},
		{"2", "SOLECTRIA", "PVI 75TL", 75.0, "c6deb7a4-5e0b-4d9e-8b4f-d051b9baccab", "Ground"},
		{"3", "SOLECTRIA", "PVI 75TL", 75.0, "7f3c2a91-3d4e-4a1b-9c5e-d051b9baccab", "Ground"},
		{"4", "SOLECTRIA", "PVI 75TL", 75.0, "2b8f5c6d-1a2b-4c3d-8e9f-d051b9baccab", "Ground"},
		{"5", "SOLECTRIA", "PVI 75TL", 75.0, "5e4d3c2b-6f7e-4d5c-9b8a-d051b9baccab", "Ground"},
		{"6", "SOLECTRIA", "PVI 75TL", 75.0, "8a9b0c1d-2e3f-4a5b-6c7d-d051b9baccab", "Ground"},
		{"7", "SOLECTRIA", "PVI 75TL", 75.0, "1d2e3f4a-5b6c-4d7e-8f9a-d051b9baccab", "Ground"},
		{"8", "SOLECTRIA", "PVI 75TL", 75.0, "4a5b6c7d-8e9f-41a2-b3c4-d051b9baccab", "Ground"},
		{"9", "SOLECTRIA", "PVI 75TL", 75.0, "7d8e9f0a-1b2c-443d-4e5f-d051b9baccab", "Ground"},
		{"10", "SOLECTRIA", "PVI 75TL", 75.0, "0a1b2c3d-4e5f-46a7-b8c9-d051b9baccab", "Ground"},
		{"11", "SOLECTRIA", "PVI 75TL", 75.0, "3d4e5f6a-7b8c-49d0-e1f2-d051b9baccab", "Ground"},
		{"12", "SOLECTRIA", "PVI 75TL", 75.0, "6a7b8c9d-0e1f-42a3-b4c5-d051b9baccab", "Ground"},
		{"13", "SOLECTRIA", "PVI 75TL", 75.0, "9d0e1f2a-3b4c-45d6-e7f8-d051b9baccab", "Ground"},
		{"14", "SOLECTRIA", "PVI 75TL", 75.0, "2a3b4c5d-6e7f-48a9-b0c1-d051b9baccab", "Ground"},
		{"15", "SOLECTRIA", "PVI 75TL", 75.0, "5d6e7f8a-9b0c-41d2-e3f4-d051b9baccab", "Ground"},
		{"16", "SOLECTRIA", "PVI 75TL", 75.0, "8a9b0c1d-2e3f-44a5-b6c7-d051b9baccab", "Ground"},
		{"17", "SOLECTRIA", "PVI 75TL", 75.0, "1d2e3f4a-5b6c-47d8-e9f0-d051b9baccab", "Ground"},
		{"18", "SOLECTRIA", "PVI 75TL", 75.0, "4a5b6c7d-8e9f-40a1-b2c3-d051b9baccab", "Ground"},
		
		// Rooftop inverters (INV-29 through INV-46) - SOLIS brand
		{"INV-29", "SOLIS", "S5-GR3P15K", 15.0, "a1b2c3d4-e5f6-4789-a0b1-d051b9baccab", "Rooftop"},
		{"INV-30", "SOLIS", "S5-GR3P15K", 15.0, "b2c3d4e5-f6a7-4890-b1c2-d051b9baccab", "Rooftop"},
		{"INV-31", "SOLIS", "S5-GR3P15K", 15.0, "c3d4e5f6-a7b8-4901-c2d3-d051b9baccab", "Rooftop"},
		{"INV-32", "SOLIS", "S5-GR3P15K", 15.0, "d4e5f6a7-b8c9-4012-d3e4-d051b9baccab", "Rooftop"},
		{"INV-33", "SOLIS", "S5-GR3P15K", 15.0, "e5f6a7b8-c9d0-4123-e4f5-d051b9baccab", "Rooftop"},
		{"INV-34", "SOLIS", "S5-GR3P15K", 15.0, "f6a7b8c9-d0e1-4234-f5a6-d051b9baccab", "Rooftop"},
		{"INV-35", "SOLIS", "S5-GR3P15K", 15.0, "a7b8c9d0-e1f2-4345-a6b7-d051b9baccab", "Rooftop"},
		{"INV-36", "SOLIS", "S5-GR3P15K", 15.0, "b8c9d0e1-f2a3-4456-b7c8-d051b9baccab", "Rooftop"},
		{"INV-37", "SOLIS", "S5-GR3P15K", 15.0, "c9d0e1f2-a3b4-4567-c8d9-d051b9baccab", "Rooftop"},
		{"INV-38", "SOLIS", "S5-GR3P15K", 15.0, "d0e1f2a3-b4c5-4678-d9e0-d051b9baccab", "Rooftop"},
		{"INV-39", "SOLIS", "S5-GR3P15K", 15.0, "e1f2a3b4-c5d6-4789-e0f1-d051b9baccab", "Rooftop"},
		{"INV-40", "SOLIS", "S5-GR3P15K", 15.0, "f2a3b4c5-d6e7-4890-f1a2-d051b9baccab", "Rooftop"},
		{"INV-41", "SOLIS", "S5-GR3P15K", 15.0, "a3b4c5d6-e7f8-4901-a2b3-d051b9baccab", "Rooftop"},
		{"INV-42", "SOLIS", "S5-GR3P15K", 15.0, "b4c5d6e7-f8a9-4012-b3c4-d051b9baccab", "Rooftop"},
		{"INV-43", "SOLIS", "S5-GR3P15K", 15.0, "c5d6e7f8-a9b0-4123-c4d5-d051b9baccab", "Rooftop"},
		{"INV-44", "SOLIS", "S5-GR3P15K", 15.0, "d6e7f8a9-b0c1-4234-d5e6-d051b9baccab", "Rooftop"},
		{"INV-45", "SOLIS", "S5-GR3P15K", 15.0, "e7f8a9b0-c1d2-4345-e6f7-d051b9baccab", "Rooftop"},
		{"INV-46", "SOLIS", "S5-GR3P15K", 15.0, "f8a9b0c1-d2e3-4456-f7a8-d051b9baccab", "Rooftop"},
	}

	// Create inverter components
	for i, inv := range inverters {
		// Check if component already exists
		var existingComponent domain.SiteComponent
		if err := tx.Where("site_id = ? AND name = ?", siteID, inv.Name).First(&existingComponent).Error; err == nil {
			// Component already exists, skip
			continue
		}
		
		spatialID := uuid.MustParse(inv.SpatialID)
		
		// Use raw SQL to insert component (avoiding embedding field issue)
		if err := tx.Exec(`
			INSERT INTO site_components (id, site_id, component_type, name, label, spatial_id, specifications, electrical_data, physical_data, current_status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, uuid.New(), siteID, string(domain.ComponentTypeInverter), inv.Name, inv.Manufacturer+" "+inv.Model, spatialID, 
		`{"manufacturer":"`+inv.Manufacturer+`","model":"`+inv.Model+`","serial_number":"`+inv.Name+"_SN_"+inv.SpatialID[:8]+`","capacity_kw":`+fmt.Sprintf("%.1f", inv.CapacityKW)+`,"voltage":"480V","frequency":"60Hz","installation":"`+inv.Location+`","string_number":`+fmt.Sprintf("%d", i+1)+`,"combiner_box":"CB-`+inv.Location[:4]+"-"+string(rune(65+i/6))+`"}`,
		`{"max_dc_power":`+fmt.Sprintf("%.0f", inv.CapacityKW*1000)+`,"max_ac_power":`+fmt.Sprintf("%.0f", inv.CapacityKW*1000*0.95)+`,"efficiency":0.97,"mppt_channels":2,"max_dc_voltage":1000,"max_ac_current":`+fmt.Sprintf("%.2f", inv.CapacityKW*1000/480/1.732)+`}`,
		`{"area":"`+inv.Location+`","row":`+fmt.Sprintf("%d", (i/6)+1)+`,"position":`+fmt.Sprintf("%d", (i%6)+1)+`,"coordinates":[-118.2437,34.0522]}`,
		string(domain.ComponentStatusOperational), time.Now(), time.Now()).Error; err != nil {
			return err
		}
	}

	return nil
}

func populateSiteDataDown(tx *gorm.DB) error {
	siteID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	
	// Delete all components for the site
	if err := tx.Where("site_id = ?", siteID).Delete(&domain.SiteComponent{}).Error; err != nil {
		return err
	}

	// Delete the site
	if err := tx.Where("id = ?", siteID).Delete(&domain.Site{}).Error; err != nil {
		return err
	}

	return nil
}