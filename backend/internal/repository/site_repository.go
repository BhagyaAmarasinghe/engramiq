package repository

import (
	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SiteRepository interface {
	Create(site *domain.Site) error
	GetByID(id uuid.UUID) (*domain.Site, error)
	GetBySiteCode(siteCode string) (*domain.Site, error)
	List(pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.Site, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GetWithDetails(id uuid.UUID) (*domain.SiteWithDetails, error)
	// Convenience methods for API handlers
	GetSites(page, limit int) ([]*domain.Site, int64, error)
	GetSite(siteID string) (*domain.Site, error)
}

type siteRepository struct {
	*BaseRepository
}

func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *siteRepository) Create(site *domain.Site) error {
	return r.db.Create(site).Error
}

func (r *siteRepository) GetByID(id uuid.UUID) (*domain.Site, error) {
	var site domain.Site
	err := r.db.First(&site, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *siteRepository) GetBySiteCode(siteCode string) (*domain.Site, error) {
	var site domain.Site
	err := r.db.First(&site, "site_code = ?", siteCode).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *siteRepository) List(pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.Site, error) {
	var sites []*domain.Site
	
	query := r.db.Model(&domain.Site{})
	query = r.ApplyFilters(query, filters)
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.Site{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	err = query.Find(&sites).Error
	
	return sites, err
}

func (r *siteRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.Site{}).Where("id = ?", id).Updates(updates).Error
}

func (r *siteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Site{}, "id = ?", id).Error
}

func (r *siteRepository) GetWithDetails(id uuid.UUID) (*domain.SiteWithDetails, error) {
	var site domain.Site
	err := r.db.First(&site, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// Get component summary
	var componentSummary struct {
		Inverters       int64
		Combiners       int64 
		TotalComponents int64
	}

	// Count inverters
	r.db.Model(&domain.SiteComponent{}).
		Where("site_id = ? AND component_type = ?", id, "inverter").
		Count(&componentSummary.Inverters)

	// Count combiners  
	r.db.Model(&domain.SiteComponent{}).
		Where("site_id = ? AND component_type = ?", id, "combiner").
		Count(&componentSummary.Combiners)

	// Count total components
	r.db.Model(&domain.SiteComponent{}).
		Where("site_id = ?", id).
		Count(&componentSummary.TotalComponents)

	// Count recent activity (last 30 days)
	var recentActivity int64
	r.db.Model(&domain.ExtractedAction{}).
		Where("site_id = ? AND created_at > NOW() - INTERVAL '30 days'", id).
		Count(&recentActivity)

	return &domain.SiteWithDetails{
		Site: site,
		ComponentSummary: domain.SiteComponentSummary{
			Inverters:       int(componentSummary.Inverters),
			Combiners:       int(componentSummary.Combiners),
			TotalComponents: int(componentSummary.TotalComponents),
		},
		RecentActivityCount: int(recentActivity),
	}, nil
}

// Convenience methods for API handlers
func (r *siteRepository) GetSites(page, limit int) ([]*domain.Site, int64, error) {
	pagination := &domain.Pagination{
		Page:  page,
		Limit: limit,
	}
	
	sites, err := r.List(pagination, nil)
	if err != nil {
		return nil, 0, err
	}
	
	return sites, pagination.TotalItems, nil
}

func (r *siteRepository) GetSite(siteID string) (*domain.Site, error) {
	// Try to parse as UUID first
	if id, err := uuid.Parse(siteID); err == nil {
		return r.GetByID(id)
	}
	
	// If not a UUID, try as site code
	return r.GetBySiteCode(siteID)
}