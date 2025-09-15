package repository

import (
	"strings"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// BeginTx starts a new transaction
func (r *BaseRepository) BeginTx() *gorm.DB {
	return r.db.Begin()
}

// WithTx returns a new repository instance with transaction
func (r *BaseRepository) WithTx(tx *gorm.DB) *BaseRepository {
	return &BaseRepository{db: tx}
}

// BuildQuery applies pagination and filtering to queries
func (r *BaseRepository) BuildQuery(query *gorm.DB, pagination *domain.Pagination) *gorm.DB {
	if pagination.Sort != "" {
		query = query.Order(pagination.Sort)
	}
	
	if pagination.Limit > 0 {
		query = query.Offset(pagination.GetOffset()).Limit(pagination.Limit)
	}
	
	return query
}

// CountTotal counts total records for pagination
func (r *BaseRepository) CountTotal(query *gorm.DB, model interface{}) (int64, error) {
	var count int64
	err := query.Model(model).Count(&count).Error
	return count, err
}

// ApplyFilters applies common filters to queries
func (r *BaseRepository) ApplyFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for field, value := range filters {
		switch field {
		case "site_id":
			if siteID, ok := value.(uuid.UUID); ok {
				query = query.Where("site_id = ?", siteID)
			}
		case "component_type":
			if componentType, ok := value.(string); ok && componentType != "" {
				query = query.Where("component_type = ?", componentType)
			}
		case "document_type":
			if docType, ok := value.(string); ok && docType != "" {
				query = query.Where("document_type = ?", docType)
			}
		case "action_type":
			if actionType, ok := value.(string); ok && actionType != "" {
				query = query.Where("action_type = ?", actionType)
			}
		case "status":
			if status, ok := value.(string); ok && status != "" {
				query = query.Where("current_status = ? OR action_status = ?", status, status)
			}
		case "date_from":
			if date, ok := value.(string); ok && date != "" {
				query = query.Where("action_date >= ? OR document_date >= ?", date, date)
			}
		case "date_to":
			if date, ok := value.(string); ok && date != "" {
				query = query.Where("action_date <= ? OR document_date <= ?", date, date)
			}
		}
	}
	return query
}

// ApplySearch adds full-text search capabilities
func (r *BaseRepository) ApplySearch(query *gorm.DB, searchTerm string, fields ...string) *gorm.DB {
	if searchTerm == "" {
		return query
	}
	
	// Use PostgreSQL full-text search
	searchQuery := "%" + searchTerm + "%"
	
	if len(fields) == 0 {
		// Default search fields
		fields = []string{"title", "name", "description", "content"}
	}
	
	// Build OR condition for multiple fields
	conditions := make([]interface{}, 0, len(fields)*2)
	placeholders := make([]string, 0, len(fields))
	
	for _, field := range fields {
		placeholders = append(placeholders, field+" ILIKE ?")
		conditions = append(conditions, searchQuery)
	}
	
	if len(placeholders) > 0 {
		query = query.Where(strings.Join(placeholders, " OR "), conditions...)
	}
	
	return query
}