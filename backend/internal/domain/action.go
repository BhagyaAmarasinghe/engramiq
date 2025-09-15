package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/lib/pq"
)

type ActionType string

const (
	ActionTypeMaintenance    ActionType = "maintenance"
	ActionTypeReplacement    ActionType = "replacement"
	ActionTypeTroubleshoot   ActionType = "troubleshoot"
	ActionTypeInspection     ActionType = "inspection"
	ActionTypeRepair         ActionType = "repair"
	ActionTypeTesting        ActionType = "testing"
	ActionTypeInstallation   ActionType = "installation"
	ActionTypeCommissioning  ActionType = "commissioning"
	ActionTypeFaultClearing  ActionType = "fault_clearing"
	ActionTypeMonitoring     ActionType = "monitoring"
	ActionTypeCleaning       ActionType = "cleaning"
	ActionTypeOther          ActionType = "other"
)

type ActionStatus string

const (
	ActionStatusPlanned         ActionStatus = "planned"
	ActionStatusInProgress      ActionStatus = "in_progress"
	ActionStatusCompleted       ActionStatus = "completed"
	ActionStatusCancelled       ActionStatus = "cancelled"
	ActionStatusOnHold          ActionStatus = "on_hold"
	ActionStatusRequiresFollowUp ActionStatus = "requires_follow_up"
)

type ExtractedAction struct {
	ID                   uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID           uuid.UUID            `json:"document_id" gorm:"type:uuid;not null"`
	Document             *Document            `json:"document,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	SiteID               uuid.UUID            `json:"site_id" gorm:"type:uuid;not null"`
	Site                 *Site                `json:"site,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	ActionType           ActionType           `json:"action_type" gorm:"type:action_type;not null"`
	Title                string               `json:"title" gorm:"type:varchar(500);not null"`
	Description          string               `json:"description"`
	ActionDate           *time.Time           `json:"action_date"`
	StartTime            *time.Time           `json:"start_time"`
	EndTime              *time.Time           `json:"end_time"`
	DurationMinutes      int                  `json:"duration_minutes"`
	TechnicianNames      pq.StringArray       `json:"technician_names" gorm:"type:text[]"`
	WorkOrderNumber      string               `json:"work_order_number" gorm:"type:varchar(100)"`
	ActionStatus         ActionStatus         `json:"action_status" gorm:"type:action_status;default:'completed'"`
	OutcomeDescription   string               `json:"outcome_description"`
	IssuesFound          pq.StringArray       `json:"issues_found" gorm:"type:text[]"`
	FollowUpActions      pq.StringArray       `json:"follow_up_actions" gorm:"type:text[]"`
	PrimaryComponentID   *uuid.UUID           `json:"primary_component_id" gorm:"type:uuid"`
	PrimaryComponent     *SiteComponent       `json:"primary_component,omitempty"`
	Measurements         JSON                 `json:"measurements" gorm:"type:jsonb;default:'{}'"`
	FaultCodes           pq.StringArray       `json:"fault_codes" gorm:"type:text[]"`
	CaseNumbers          pq.StringArray       `json:"case_numbers" gorm:"type:text[]"`
	ExtractionConfidence float64              `json:"extraction_confidence"`
	ExtractionModel      string               `json:"extraction_model" gorm:"type:varchar(50)"`
	ExtractionMetadata   JSON                 `json:"extraction_metadata" gorm:"type:jsonb;default:'{}'"`
	Embedding            pgvector.Vector      `json:"-" gorm:"type:vector(1536)"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

func (ExtractedAction) TableName() string {
	return "extracted_actions"
}

type ActionComponent struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ActionID         uuid.UUID      `json:"action_id" gorm:"type:uuid;not null"`
	ComponentID      uuid.UUID      `json:"component_id" gorm:"type:uuid;not null"`
	InvolvementType  string         `json:"involvement_type" gorm:"type:varchar(50)"`
	MentionText      string         `json:"mention_text"`
	ConfidenceScore  float64        `json:"confidence_score"`
	CreatedAt        time.Time      `json:"created_at"`
}

func (ActionComponent) TableName() string {
	return "action_components"
}

// ActionWithComponents includes related components
type ActionWithComponents struct {
	ExtractedAction
	RelatedComponents []ActionComponentDetail `json:"related_components,omitempty"`
}

type ActionComponentDetail struct {
	ComponentID     uuid.UUID  `json:"component_id"`
	Component       SiteComponent `json:"component"`
	InvolvementType string     `json:"involvement_type"`
	ConfidenceScore float64    `json:"confidence_score"`
}