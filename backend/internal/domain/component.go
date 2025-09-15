package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type ComponentType string

const (
	ComponentTypeInverter    ComponentType = "inverter"
	ComponentTypeCombiner    ComponentType = "combiner"
	ComponentTypePanel       ComponentType = "panel"
	ComponentTypeTransformer ComponentType = "transformer"
	ComponentTypeMeter       ComponentType = "meter"
	ComponentTypeSwitchgear  ComponentType = "switchgear"
	ComponentTypeMonitoring  ComponentType = "monitoring"
	ComponentTypeOther       ComponentType = "other"
)

type ComponentStatus string

const (
	ComponentStatusOperational ComponentStatus = "operational"
	ComponentStatusFault      ComponentStatus = "fault"
	ComponentStatusMaintenance ComponentStatus = "maintenance"
	ComponentStatusOffline    ComponentStatus = "offline"
)

type SiteComponent struct {
	ID                  uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID              uuid.UUID            `json:"site_id" gorm:"type:uuid;not null"`
	Site                *Site                `json:"site,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	ExternalID          string               `json:"external_id" gorm:"type:varchar(255)"`
	ComponentType       ComponentType        `json:"component_type" gorm:"type:component_type;not null"`
	Name                string               `json:"name" gorm:"type:varchar(255);not null"`
	Label               string               `json:"label" gorm:"type:varchar(255)"`
	Level               int                  `json:"level" gorm:"default:0"`
	GroupName           string               `json:"group_name" gorm:"type:varchar(255)"`
	Specifications      JSON                 `json:"specifications" gorm:"type:jsonb;default:'{}'"`
	ElectricalData      JSON                 `json:"electrical_data" gorm:"type:jsonb;default:'{}'"`
	PhysicalData        JSON                 `json:"physical_data" gorm:"type:jsonb;default:'{}'"`
	DrawingTitle        string               `json:"drawing_title" gorm:"type:varchar(500)"`
	DrawingNumber       string               `json:"drawing_number" gorm:"type:varchar(100)"`
	Revision            string               `json:"revision" gorm:"type:varchar(50)"`
	RevisionDate        *time.Time           `json:"revision_date"`
	SpatialID           *uuid.UUID           `json:"spatial_id" gorm:"type:uuid"`
	Coordinates         *Point               `json:"coordinates" gorm:"type:varchar(100)"`
	Embedding           pgvector.Vector      `json:"-" gorm:"type:vector(1536)"`
	CurrentStatus       ComponentStatus      `json:"current_status" gorm:"type:varchar(50);default:'operational'"`
	LastMaintenanceDate *time.Time           `json:"last_maintenance_date"`
	NextMaintenanceDate *time.Time           `json:"next_maintenance_date"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	DeletedAt           gorm.DeletedAt       `json:"deleted_at,omitempty" gorm:"index"`
}

func (SiteComponent) TableName() string {
	return "site_components"
}

type ComponentRelationshipType string

const (
	RelationshipConnectsTo   ComponentRelationshipType = "connects_to"
	RelationshipPowers       ComponentRelationshipType = "powers"
	RelationshipControls     ComponentRelationshipType = "controls"
	RelationshipMonitors     ComponentRelationshipType = "monitors"
	RelationshipParentChild  ComponentRelationshipType = "parent_child"
	RelationshipSameString   ComponentRelationshipType = "same_string"
	RelationshipSameCombiner ComponentRelationshipType = "same_combiner"
)

type ComponentRelationship struct {
	ID                 uuid.UUID                 `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ParentComponentID  uuid.UUID                 `json:"parent_component_id" gorm:"type:uuid;not null"`
	ChildComponentID   uuid.UUID                 `json:"child_component_id" gorm:"type:uuid;not null"`
	RelationshipType   ComponentRelationshipType `json:"relationship_type" gorm:"type:relationship_type;not null"`
	RelationshipData   JSON                      `json:"relationship_data" gorm:"type:jsonb;default:'{}'"`
	CreatedAt          time.Time                 `json:"created_at"`
}

func (ComponentRelationship) TableName() string {
	return "component_relationships"
}

// ComponentWithTimeline includes recent timeline events
type ComponentWithTimeline struct {
	SiteComponent
	RecentEvents []SiteEvent `json:"recent_events,omitempty"`
}