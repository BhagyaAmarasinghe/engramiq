package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type EventType string

const (
	EventTypeMaintenanceScheduled  EventType = "maintenance_scheduled"
	EventTypeMaintenanceCompleted  EventType = "maintenance_completed"
	EventTypeFaultOccurred        EventType = "fault_occurred"
	EventTypeFaultCleared         EventType = "fault_cleared"
	EventTypeReplacementScheduled EventType = "replacement_scheduled"
	EventTypeReplacementCompleted EventType = "replacement_completed"
	EventTypeInspectionScheduled  EventType = "inspection_scheduled"
	EventTypeInspectionCompleted  EventType = "inspection_completed"
	EventTypeWarrantyClaim        EventType = "warranty_claim"
	EventTypePerformanceAlert     EventType = "performance_alert"
	EventTypeContractMilestone    EventType = "contract_milestone"
	EventTypeOther               EventType = "other"
)

type EventPriority string

const (
	EventPriorityLow      EventPriority = "low"
	EventPriorityMedium   EventPriority = "medium"
	EventPriorityHigh     EventPriority = "high"
	EventPriorityCritical EventPriority = "critical"
)

type SiteEvent struct {
	ID                     uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID                 uuid.UUID      `json:"site_id" gorm:"type:uuid;not null"`
	Site                   *Site          `json:"site,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	ActionID               *uuid.UUID     `json:"action_id" gorm:"type:uuid"`
	Action                 *ExtractedAction `json:"action,omitempty" gorm:"constraint:OnDelete:SET NULL"`
	EventType              EventType      `json:"event_type" gorm:"type:event_type;not null"`
	Title                  string         `json:"title" gorm:"type:varchar(500);not null"`
	Description            string         `json:"description"`
	StartTime              time.Time      `json:"start_time" gorm:"not null"`
	EndTime                *time.Time     `json:"end_time"`
	IsAllDay               bool           `json:"is_all_day" gorm:"default:false"`
	IsFuture               bool           `json:"is_future" gorm:"default:false"`
	Priority               EventPriority  `json:"priority" gorm:"type:event_priority;default:'medium'"`
	Status                 string         `json:"status" gorm:"type:varchar(50);default:'active'"`
	PrimaryComponentID     *uuid.UUID     `json:"primary_component_id" gorm:"type:uuid"`
	PrimaryComponent       *SiteComponent `json:"primary_component,omitempty"`
	AffectedComponentIDs   pq.StringArray `json:"affected_component_ids" gorm:"type:uuid[]"`
	WorkOrderNumber        string         `json:"work_order_number" gorm:"type:varchar(100)"`
	TechnicianAssigned     string         `json:"technician_assigned" gorm:"type:varchar(255)"`
	EstimatedDurationHours float64        `json:"estimated_duration_hours"`
	SourceDocumentID       *uuid.UUID     `json:"source_document_id" gorm:"type:uuid"`
	SourceDocument         *Document      `json:"source_document,omitempty"`
	EventMetadata          JSON           `json:"event_metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

func (SiteEvent) TableName() string {
	return "site_events"
}

// TimelineResponse structures for API responses
type TimelineEvent struct {
	ID                     uuid.UUID         `json:"id"`
	Title                  string            `json:"title"`
	Description            string            `json:"description"`
	StartTime              time.Time         `json:"start_time"`
	EndTime                *time.Time        `json:"end_time,omitempty"`
	EventType              EventType         `json:"event_type"`
	Priority               EventPriority     `json:"priority"`
	IsFuture               bool              `json:"is_future"`
	Component              *ComponentSummary `json:"component,omitempty"`
	Sources                []DocumentSource  `json:"sources,omitempty"`
	WorkOrderNumber        string            `json:"work_order_number,omitempty"`
	TechnicianAssigned     string            `json:"technician_assigned,omitempty"`
	FollowUpActions        []string          `json:"follow_up_actions,omitempty"`
	Metadata               JSON              `json:"metadata"`
}

type ComponentSummary struct {
	ID            uuid.UUID     `json:"id"`
	ExternalID    string        `json:"external_id"`
	Name          string        `json:"name"`
	ComponentType ComponentType `json:"component_type"`
}

type DocumentSource struct {
	DocumentID    uuid.UUID `json:"document_id"`
	DocumentTitle string    `json:"document_title"`
}

type TimelineSummary struct {
	TotalEvents       int `json:"total_events"`
	MaintenanceEvents int `json:"maintenance_events"`
	FaultEvents       int `json:"fault_events"`
	UpcomingEvents    int `json:"upcoming_events"`
	CriticalEvents    int `json:"critical_events"`
}