package repository

import (
	"fmt"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type ActionRepository interface {
	Create(action *domain.ExtractedAction) error
	CreateWithComponents(action *domain.ExtractedAction, components []domain.ActionComponent) error
	GetByID(id uuid.UUID) (*domain.ActionWithComponents, error)
	ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.ExtractedAction, error)
	ListByComponent(componentID uuid.UUID, pagination *domain.Pagination) ([]*domain.ExtractedAction, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	SearchSemantic(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.ExtractedAction, error)
	GetByWorkOrderNumber(workOrder string) ([]*domain.ExtractedAction, error)
	GetMaintenanceHistory(componentID uuid.UUID, limit int) ([]*domain.ExtractedAction, error)
	GetByDateRange(siteID uuid.UUID, startDate, endDate time.Time) ([]*domain.ExtractedAction, error)
}

type actionRepository struct {
	*BaseRepository
}

func NewActionRepository(db *gorm.DB) ActionRepository {
	return &actionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *actionRepository) Create(action *domain.ExtractedAction) error {
	return r.db.Create(action).Error
}

func (r *actionRepository) CreateWithComponents(action *domain.ExtractedAction, components []domain.ActionComponent) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the action
		if err := tx.Create(action).Error; err != nil {
			return err
		}
		
		// Create component relationships
		for i := range components {
			components[i].ActionID = action.ID
		}
		
		if len(components) > 0 {
			if err := tx.Create(&components).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
}

func (r *actionRepository) GetByID(id uuid.UUID) (*domain.ActionWithComponents, error) {
	var action domain.ExtractedAction
	err := r.db.Preload("Document").
		Preload("Site").
		Preload("PrimaryComponent").
		First(&action, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// Get related components
	var actionComponents []domain.ActionComponent
	err = r.db.Where("action_id = ?", id).Find(&actionComponents).Error
	if err != nil {
		return nil, err
	}

	// Build the response with component details
	relatedComponents := make([]domain.ActionComponentDetail, len(actionComponents))
	for i, ac := range actionComponents {
		var component domain.SiteComponent
		r.db.First(&component, "id = ?", ac.ComponentID)
		
		relatedComponents[i] = domain.ActionComponentDetail{
			ComponentID:     ac.ComponentID,
			Component:       component,
			InvolvementType: ac.InvolvementType,
			ConfidenceScore: ac.ConfidenceScore,
		}
	}

	return &domain.ActionWithComponents{
		ExtractedAction:   action,
		RelatedComponents: relatedComponents,
	}, nil
}

func (r *actionRepository) ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	query := r.db.Model(&domain.ExtractedAction{}).
		Preload("PrimaryComponent").
		Where("site_id = ?", siteID)
	
	query = r.ApplyFilters(query, filters)
	
	// Additional specific filters
	if componentID, ok := filters["component_id"].(uuid.UUID); ok {
		// Join with action_components table to find actions related to specific component
		query = query.Joins("LEFT JOIN action_components ac ON extracted_actions.id = ac.action_id").
			Where("extracted_actions.primary_component_id = ? OR ac.component_id = ?", componentID, componentID)
	}
	
	if workOrder, ok := filters["work_order_number"].(string); ok && workOrder != "" {
		query = query.Where("work_order_number = ?", workOrder)
	}
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.ExtractedAction{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	err = query.Find(&actions).Error
	
	return actions, err
}

func (r *actionRepository) ListByComponent(componentID uuid.UUID, pagination *domain.Pagination) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	query := r.db.Model(&domain.ExtractedAction{}).
		Joins("LEFT JOIN action_components ac ON extracted_actions.id = ac.action_id").
		Where("extracted_actions.primary_component_id = ? OR ac.component_id = ?", componentID, componentID).
		Order("action_date DESC, created_at DESC")
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.ExtractedAction{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination
	query = r.BuildQuery(query, pagination)
	err = query.Find(&actions).Error
	
	return actions, err
}

func (r *actionRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.ExtractedAction{}).Where("id = ?", id).Updates(updates).Error
}

func (r *actionRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete related action_components first
		if err := tx.Delete(&domain.ActionComponent{}, "action_id = ?", id).Error; err != nil {
			return err
		}
		
		// Delete the action
		return tx.Delete(&domain.ExtractedAction{}, "id = ?", id).Error
	})
}

func (r *actionRepository) SearchSemantic(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	err := r.db.Preload("PrimaryComponent").
		Where("site_id = ?", siteID).
		Where("embedding <=> ? < ?", embedding, threshold).
		Order(fmt.Sprintf("embedding <=> '%v'", embedding)).
		Limit(limit).
		Find(&actions).Error
	
	return actions, err
}

func (r *actionRepository) GetByWorkOrderNumber(workOrder string) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	err := r.db.Preload("PrimaryComponent").
		Where("work_order_number = ?", workOrder).
		Order("action_date DESC").
		Find(&actions).Error
	
	return actions, err
}

func (r *actionRepository) GetMaintenanceHistory(componentID uuid.UUID, limit int) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	err := r.db.Preload("Document").
		Joins("LEFT JOIN action_components ac ON extracted_actions.id = ac.action_id").
		Where("extracted_actions.primary_component_id = ? OR ac.component_id = ?", componentID, componentID).
		Where("action_type IN (?)", []string{"maintenance", "replacement", "repair", "troubleshoot"}).
		Order("action_date DESC, created_at DESC").
		Limit(limit).
		Find(&actions).Error
	
	return actions, err
}

func (r *actionRepository) GetByDateRange(siteID uuid.UUID, startDate, endDate time.Time) ([]*domain.ExtractedAction, error) {
	var actions []*domain.ExtractedAction
	
	err := r.db.Preload("PrimaryComponent").
		Where("site_id = ?", siteID).
		Where("action_date BETWEEN ? AND ?", startDate, endDate).
		Order("action_date ASC").
		Find(&actions).Error
	
	return actions, err
}