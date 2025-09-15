package repository

import (
	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ComponentRepository interface {
	Create(component *domain.SiteComponent) error
	GetByID(id uuid.UUID) (*domain.SiteComponent, error)
	GetByExternalID(siteID uuid.UUID, externalID string) (*domain.SiteComponent, error)
	ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.SiteComponent, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GetHierarchy(siteID uuid.UUID) ([]*domain.SiteComponent, error)
	FindBySpecification(siteID uuid.UUID, key string, value string) ([]*domain.SiteComponent, error)
	BulkCreate(components []*domain.SiteComponent) error
}

type componentRepository struct {
	*BaseRepository
}

func NewComponentRepository(db *gorm.DB) ComponentRepository {
	return &componentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *componentRepository) Create(component *domain.SiteComponent) error {
	return r.db.Create(component).Error
}

func (r *componentRepository) GetByID(id uuid.UUID) (*domain.SiteComponent, error) {
	var component domain.SiteComponent
	err := r.db.Preload("Site").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *componentRepository) GetByExternalID(siteID uuid.UUID, externalID string) (*domain.SiteComponent, error) {
	var component domain.SiteComponent
	err := r.db.Preload("Site").
		First(&component, "site_id = ? AND external_id = ?", siteID, externalID).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *componentRepository) ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.SiteComponent, error) {
	var components []*domain.SiteComponent
	
	query := r.db.Model(&domain.SiteComponent{}).Where("site_id = ?", siteID)
	query = r.ApplyFilters(query, filters)
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.SiteComponent{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	err = query.Find(&components).Error
	
	return components, err
}

func (r *componentRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.SiteComponent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *componentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SiteComponent{}, "id = ?", id).Error
}

func (r *componentRepository) GetHierarchy(siteID uuid.UUID) ([]*domain.SiteComponent, error) {
	var components []*domain.SiteComponent
	err := r.db.Where("site_id = ?", siteID).
		Order("level ASC, sort_order ASC, external_id ASC").
		Find(&components).Error
	return components, err
}

func (r *componentRepository) FindBySpecification(siteID uuid.UUID, key string, value string) ([]*domain.SiteComponent, error) {
	var components []*domain.SiteComponent
	
	// Use JSONB query to search in specifications
	err := r.db.Where("site_id = ? AND specifications->>? = ?", siteID, key, value).
		Find(&components).Error
	
	return components, err
}

func (r *componentRepository) BulkCreate(components []*domain.SiteComponent) error {
	// Use batch insert for better performance
	batchSize := 100
	for i := 0; i < len(components); i += batchSize {
		end := i + batchSize
		if end > len(components) {
			end = len(components)
		}
		
		if err := r.db.Create(components[i:end]).Error; err != nil {
			return err
		}
	}
	return nil
}