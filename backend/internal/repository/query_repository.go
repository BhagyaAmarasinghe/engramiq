package repository

import (
	"fmt"
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type QueryRepository interface {
	Create(query *domain.UserQuery) error
	GetByID(id uuid.UUID) (*domain.UserQuery, error)
	ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.UserQuery, error)
	ListByUser(userID uuid.UUID, pagination *domain.Pagination) ([]*domain.UserQuery, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	UpdateResults(id uuid.UUID, results domain.JSON, resultCount int) error
	GetRecentQueries(siteID uuid.UUID, limit int) ([]*domain.UserQuery, error)
	SearchSimilarQueries(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.UserQuery, error)
	GetQueryAnalytics(siteID uuid.UUID, startDate, endDate time.Time) (*domain.QueryAnalytics, error)
}

type queryRepository struct {
	*BaseRepository
}

func NewQueryRepository(db *gorm.DB) QueryRepository {
	return &queryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *queryRepository) Create(query *domain.UserQuery) error {
	return r.db.Create(query).Error
}

func (r *queryRepository) GetByID(id uuid.UUID) (*domain.UserQuery, error) {
	var query domain.UserQuery
	err := r.db.Preload("Site").First(&query, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &query, nil
}

func (r *queryRepository) ListBySite(siteID uuid.UUID, pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.UserQuery, error) {
	var queries []*domain.UserQuery
	
	query := r.db.Model(&domain.UserQuery{}).Where("site_id = ?", siteID)
	query = r.ApplyFilters(query, filters)
	
	// Additional query-specific filters
	if userID, ok := filters["user_id"].(uuid.UUID); ok {
		query = query.Where("user_id = ?", userID)
	}
	
	if queryType, ok := filters["query_type"].(domain.QueryType); ok {
		query = query.Where("query_type = ?", queryType)
	}
	
	if hasResults, ok := filters["has_results"].(bool); ok {
		if hasResults {
			query = query.Where("result_count > 0")
		} else {
			query = query.Where("result_count = 0 OR result_count IS NULL")
		}
	}
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.UserQuery{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	if pagination.Sort == "" {
		query = query.Order("created_at DESC")
	}
	
	err = query.Find(&queries).Error
	
	return queries, err
}

func (r *queryRepository) ListByUser(userID uuid.UUID, pagination *domain.Pagination) ([]*domain.UserQuery, error) {
	var queries []*domain.UserQuery
	
	query := r.db.Model(&domain.UserQuery{}).
		Where("user_id = ?", userID).
		Preload("Site")
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.UserQuery{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	if pagination.Sort == "" {
		query = query.Order("created_at DESC")
	}
	
	err = query.Find(&queries).Error
	
	return queries, err
}

func (r *queryRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.UserQuery{}).Where("id = ?", id).Updates(updates).Error
}

func (r *queryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.UserQuery{}, "id = ?", id).Error
}

func (r *queryRepository) UpdateResults(id uuid.UUID, results domain.JSON, resultCount int) error {
	updates := map[string]interface{}{
		"results":      results,
		"result_count": resultCount,
		"processed_at": time.Now(),
	}
	
	return r.db.Model(&domain.UserQuery{}).Where("id = ?", id).Updates(updates).Error
}

func (r *queryRepository) GetRecentQueries(siteID uuid.UUID, limit int) ([]*domain.UserQuery, error) {
	var queries []*domain.UserQuery
	
	err := r.db.Where("site_id = ?", siteID).
		Where("result_count > 0").
		Order("created_at DESC").
		Limit(limit).
		Find(&queries).Error
	
	return queries, err
}

func (r *queryRepository) SearchSimilarQueries(siteID uuid.UUID, embedding pgvector.Vector, limit int, threshold float64) ([]*domain.UserQuery, error) {
	var queries []*domain.UserQuery
	
	err := r.db.Where("site_id = ?", siteID).
		Where("embedding <=> ? < ?", embedding, threshold).
		Order(fmt.Sprintf("embedding <=> '%v'", embedding)).
		Limit(limit).
		Find(&queries).Error
	
	return queries, err
}

func (r *queryRepository) GetQueryAnalytics(siteID uuid.UUID, startDate, endDate time.Time) (*domain.QueryAnalytics, error) {
	var analytics domain.QueryAnalytics
	
	// Total queries count
	r.db.Model(&domain.UserQuery{}).
		Where("site_id = ? AND created_at BETWEEN ? AND ?", siteID, startDate, endDate).
		Count(&analytics.TotalQueries)
	
	// Successful queries (with results)
	r.db.Model(&domain.UserQuery{}).
		Where("site_id = ? AND created_at BETWEEN ? AND ? AND result_count > 0", siteID, startDate, endDate).
		Count(&analytics.SuccessfulQueries)
	
	// Calculate success rate
	if analytics.TotalQueries > 0 {
		analytics.SuccessRate = float64(analytics.SuccessfulQueries) / float64(analytics.TotalQueries) * 100
	}
	
	// Average response time
	var avgResponseTime struct {
		AvgTime *float64
	}
	r.db.Model(&domain.UserQuery{}).
		Select("AVG(EXTRACT(EPOCH FROM (processed_at - created_at))) as avg_time").
		Where("site_id = ? AND created_at BETWEEN ? AND ? AND processed_at IS NOT NULL", siteID, startDate, endDate).
		Scan(&avgResponseTime)
	
	if avgResponseTime.AvgTime != nil {
		analytics.AverageResponseTime = *avgResponseTime.AvgTime
	}
	
	// Most common query types
	var queryTypeStats []struct {
		QueryType string `json:"query_type"`
		Count     int64  `json:"count"`
	}
	
	r.db.Model(&domain.UserQuery{}).
		Select("query_type, COUNT(*) as count").
		Where("site_id = ? AND created_at BETWEEN ? AND ?", siteID, startDate, endDate).
		Group("query_type").
		Order("count DESC").
		Limit(10).
		Scan(&queryTypeStats)
	
	analytics.QueryTypeBreakdown = make(map[string]int64)
	for _, stat := range queryTypeStats {
		analytics.QueryTypeBreakdown[stat.QueryType] = stat.Count
	}
	
	return &analytics, nil
}