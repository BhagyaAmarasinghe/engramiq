package repository

import (
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *domain.SiteEvent) error
	GetByID(id uuid.UUID) (*domain.SiteEvent, error)
	ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.SiteEvent, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GetTimelineEvents(siteID uuid.UUID, startDate, endDate time.Time, eventTypes []domain.EventType) ([]*domain.SiteEvent, error)
	GetByEntityReference(entityType string, entityID uuid.UUID) ([]*domain.SiteEvent, error)
	MarkAsProcessed(id uuid.UUID) error
	GetPendingEvents(limit int) ([]*domain.SiteEvent, error)
}

type eventRepository struct {
	*BaseRepository
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *eventRepository) Create(event *domain.SiteEvent) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) GetByID(id uuid.UUID) (*domain.SiteEvent, error) {
	var event domain.SiteEvent
	err := r.db.Preload("Site").First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.SiteEvent, error) {
	var events []*domain.SiteEvent
	
	query := r.db.Model(&domain.SiteEvent{}).Where("site_id = ?", siteID)
	query = r.ApplyFilters(query, filters)
	
	// Additional event-specific filters
	if eventType, ok := filters["event_type"].(domain.EventType); ok {
		query = query.Where("event_type = ?", eventType)
	}
	
	if severity, ok := filters["severity"].(string); ok {
		query = query.Where("severity = ?", severity)
	}
	
	if processed, ok := filters["is_processed"].(bool); ok {
		query = query.Where("is_processed = ?", processed)
	}
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.SiteEvent{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	if pagination.Sort == "" {
		query = query.Order("event_timestamp DESC, created_at DESC")
	}
	
	err = query.Find(&events).Error
	
	return events, err
}

func (r *eventRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.SiteEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *eventRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.SiteEvent{}, "id = ?", id).Error
}

func (r *eventRepository) GetTimelineEvents(siteID uuid.UUID, startDate, endDate time.Time, eventTypes []domain.EventType) ([]*domain.SiteEvent, error) {
	var events []*domain.SiteEvent
	
	query := r.db.Where("site_id = ?", siteID).
		Where("event_timestamp BETWEEN ? AND ?", startDate, endDate)
	
	if len(eventTypes) > 0 {
		query = query.Where("event_type IN ?", eventTypes)
	}
	
	err := query.Order("event_timestamp ASC").Find(&events).Error
	
	return events, err
}

func (r *eventRepository) GetByEntityReference(entityType string, entityID uuid.UUID) ([]*domain.SiteEvent, error) {
	var events []*domain.SiteEvent
	
	err := r.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("event_timestamp DESC").
		Find(&events).Error
	
	return events, err
}

func (r *eventRepository) MarkAsProcessed(id uuid.UUID) error {
	updates := map[string]interface{}{
		"is_processed": true,
		"processed_at": time.Now(),
	}
	
	return r.db.Model(&domain.SiteEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *eventRepository) GetPendingEvents(limit int) ([]*domain.SiteEvent, error) {
	var events []*domain.SiteEvent
	
	err := r.db.Where("is_processed = ?", false).
		Order("event_timestamp ASC").
		Limit(limit).
		Find(&events).Error
	
	return events, err
}