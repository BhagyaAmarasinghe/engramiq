package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type DocumentType string

const (
	DocumentTypeFieldServiceReport DocumentType = "field_service_report"
	DocumentTypeEmail             DocumentType = "email"
	DocumentTypeMeetingTranscript DocumentType = "meeting_transcript"
	DocumentTypeWorkOrder         DocumentType = "work_order"
	DocumentTypeInspectionReport  DocumentType = "inspection_report"
	DocumentTypeWarrantyClaim     DocumentType = "warranty_claim"
	DocumentTypeContract          DocumentType = "contract"
	DocumentTypeManual           DocumentType = "manual"
	DocumentTypeDrawing          DocumentType = "drawing"
	DocumentTypeOther            DocumentType = "other"
)

type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

type Document struct {
	ID                     uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID                 uuid.UUID        `json:"site_id" gorm:"type:uuid;not null"`
	Site                   *Site            `json:"site,omitempty" gorm:"constraint:OnDelete:CASCADE"`
	DocumentType           DocumentType     `json:"document_type" gorm:"type:document_type;not null"`
	Title                  string           `json:"title" gorm:"type:varchar(500)"`
	SourceType             string           `json:"source_type" gorm:"type:varchar(100)"`
	SourceIdentifier       string           `json:"source_identifier" gorm:"type:varchar(255)"`
	RawContent             string           `json:"raw_content,omitempty"`
	ProcessedContent       string           `json:"processed_content,omitempty"`
	ContentHash            string           `json:"content_hash" gorm:"type:varchar(64)"`
	OriginalFilename       string           `json:"original_filename" gorm:"type:varchar(500)"`
	FileSize               int64            `json:"file_size"`
	MimeType               string           `json:"mime_type" gorm:"type:varchar(100)"`
	StoragePath            string           `json:"storage_path" gorm:"type:varchar(1000)"`
	ProcessingStatus       ProcessingStatus `json:"processing_status" gorm:"type:varchar(50);default:'pending'"`
	ProcessingStartedAt    *time.Time       `json:"processing_started_at"`
	ProcessingCompletedAt  *time.Time       `json:"processing_completed_at"`
	DocumentDate           *time.Time       `json:"document_date"`
	AuthorName             string           `json:"author_name" gorm:"type:varchar(255)"`
	AuthorEmail            string           `json:"author_email" gorm:"type:varchar(255)"`
	DocumentMetadata       JSON             `json:"document_metadata" gorm:"type:jsonb;default:'{}'"`
	Embedding              pgvector.Vector  `json:"-" gorm:"type:vector(1536)"`
	ContentVector          string           `json:"-" gorm:"type:tsvector"`
	UploadedBy             *uuid.UUID       `json:"uploaded_by" gorm:"type:uuid"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	DeletedAt              gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index"`
}

func (Document) TableName() string {
	return "documents"
}

type DocumentWithStats struct {
	Document
	ExtractedActionsCount int `json:"extracted_actions_count"`
}

type DocumentProcessingResult struct {
	DocumentID      uuid.UUID         `json:"document_id"`
	Status          ProcessingStatus  `json:"processing_status"`
	ExtractedActions []ExtractedAction `json:"extracted_actions,omitempty"`
	Errors          []string          `json:"errors,omitempty"`
	ProcessingTime  time.Duration     `json:"processing_time"`
}