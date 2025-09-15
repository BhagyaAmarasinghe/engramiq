package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Site struct {
	ID                 uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteCode          string          `json:"site_code" gorm:"type:varchar(50);unique;not null"`
	Name              string          `json:"name" gorm:"type:varchar(255);not null"`
	Address           string          `json:"address"`
	Country           string          `json:"country" gorm:"type:varchar(2);default:'US'"`
	TotalCapacityKW   float64         `json:"total_capacity_kw"`
	NumberOfInverters int             `json:"number_of_inverters"`
	InstallationDate  *time.Time      `json:"installation_date"`
	SiteMetadata      JSON            `json:"site_metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	DeletedAt         gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

func (Site) TableName() string {
	return "sites"
}

type SiteComponentSummary struct {
	Inverters        int `json:"inverters"`
	Combiners        int `json:"combiners"`
	TotalComponents  int `json:"total_components"`
}

type SiteWithDetails struct {
	Site
	ComponentSummary     SiteComponentSummary `json:"component_summary"`
	RecentActivityCount  int                  `json:"recent_activity_count"`
}