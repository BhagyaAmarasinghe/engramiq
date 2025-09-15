package repository

import (
	"fmt"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type DocumentRepository interface {
	Create(document *domain.Document) error
	GetByID(id uuid.UUID) (*domain.Document, error)
	ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.DocumentWithStats, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GetByContentHash(hash string) (*domain.Document, error)
	UpdateProcessingStatus(id uuid.UUID, status domain.ProcessingStatus) error
	SearchFullText(siteID uuid.UUID, query string, limit int) ([]*domain.Document, error)
	SearchSemantic(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.Document, error)
	GetPendingProcessing(limit int) ([]*domain.Document, error)
}

type documentRepository struct {
	*BaseRepository
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *documentRepository) Create(document *domain.Document) error {
	return r.db.Create(document).Error
}

func (r *documentRepository) GetByID(id uuid.UUID) (*domain.Document, error) {
	var document domain.Document
	err := r.db.Preload("Site").First(&document, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.DocumentWithStats, error) {
	var documents []*domain.DocumentWithStats
	
	query := `
		SELECT 
			d.*,
			COUNT(ea.id) as extracted_actions_count
		FROM documents d
		LEFT JOIN extracted_actions ea ON d.id = ea.document_id
		WHERE d.site_id = ?
	`
	
	args := []interface{}{siteID}
	
	// Add filters
	if docType, ok := filters["document_type"].(string); ok && docType != "" {
		query += " AND d.document_type = ?"
		args = append(args, docType)
	}
	
	if status, ok := filters["processing_status"].(string); ok && status != "" {
		query += " AND d.processing_status = ?"
		args = append(args, status)
	}
	
	query += " GROUP BY d.id"
	
	// Add ordering and pagination
	if pagination.Sort != "" {
		query += fmt.Sprintf(" ORDER BY d.%s", pagination.Sort)
	} else {
		query += " ORDER BY d.created_at DESC"
	}
	
	if pagination.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pagination.Limit, pagination.GetOffset())
	}
	
	err := r.db.Raw(query, args...).Scan(&documents).Error
	
	// Count total for pagination
	countQuery := "SELECT COUNT(DISTINCT d.id) FROM documents d WHERE d.site_id = ?"
	countArgs := []interface{}{siteID}
	
	if docType, ok := filters["document_type"].(string); ok && docType != "" {
		countQuery += " AND d.document_type = ?"
		countArgs = append(countArgs, docType)
	}
	
	var count int64
	r.db.Raw(countQuery, countArgs...).Scan(&count)
	pagination.SetTotalPages(count)
	
	return documents, err
}

func (r *documentRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.Document{}).Where("id = ?", id).Updates(updates).Error
}

func (r *documentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Document{}, "id = ?", id).Error
}

func (r *documentRepository) GetByContentHash(hash string) (*domain.Document, error) {
	var document domain.Document
	err := r.db.First(&document, "content_hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) UpdateProcessingStatus(id uuid.UUID, status domain.ProcessingStatus) error {
	updates := map[string]interface{}{
		"processing_status": status,
	}
	
	if status == domain.ProcessingStatusProcessing {
		updates["processing_started_at"] = "NOW()"
	} else if status == domain.ProcessingStatusCompleted || status == domain.ProcessingStatusFailed {
		updates["processing_completed_at"] = "NOW()"
	}
	
	return r.db.Model(&domain.Document{}).Where("id = ?", id).Updates(updates).Error
}

func (r *documentRepository) SearchFullText(siteID uuid.UUID, query string, limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	
	// Use PostgreSQL full-text search with computed tsvector
	err := r.db.Where("site_id = ?", siteID).
		Where("to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(processed_content, '')) @@ plainto_tsquery('english', ?)", query).
		Order(fmt.Sprintf("ts_rank(to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(processed_content, '')), plainto_tsquery('english', '%s')) DESC", query)).
		Limit(limit).
		Find(&documents).Error
	
	return documents, err
}

func (r *documentRepository) SearchSemantic(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.Document, error) {
	var documents []*domain.Document
	
	// Use pgvector for semantic similarity search
	// Explicitly select all fields including content fields
	err := r.db.Select("*").
		Where("site_id = ?", siteID).
		Where("embedding <=> ? < ?", embedding, threshold).
		Order(fmt.Sprintf("embedding <=> '%v'", embedding)).
		Limit(limit).
		Find(&documents).Error
	
	return documents, err
}

func (r *documentRepository) GetPendingProcessing(limit int) ([]*domain.Document, error) {
	var documents []*domain.Document
	
	err := r.db.Where("processing_status = ?", domain.ProcessingStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&documents).Error
	
	return documents, err
}