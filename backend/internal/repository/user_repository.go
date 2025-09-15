package repository

import (
	"time"

	"github.com/engramiq/engramiq-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	List(pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.User, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	UpdateLastLogin(id uuid.UUID) error
	CreateSession(session *domain.UserSession) error
	GetSessionByToken(token string) (*domain.UserSession, error)
	UpdateSession(id uuid.UUID, updates map[string]interface{}) error
	DeleteSession(id uuid.UUID) error
	DeleteExpiredSessions() error
}

type userRepository struct {
	*BaseRepository
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(pagination *domain.Pagination, filters map[string]interface{}) ([]*domain.User, error) {
	var users []*domain.User
	
	query := r.db.Model(&domain.User{})
	query = r.ApplyFilters(query, filters)
	
	// Additional user-specific filters
	if role, ok := filters["role"].(domain.UserRole); ok {
		query = query.Where("role = ?", role)
	}
	
	if isActive, ok := filters["is_active"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}
	
	// Search by name or email
	if search, ok := filters["search"].(string); ok && search != "" {
		query = r.ApplySearch(query, search, "full_name", "email")
	}
	
	// Count total for pagination
	count, err := r.CountTotal(query, &domain.User{})
	if err != nil {
		return nil, err
	}
	pagination.SetTotalPages(count)
	
	// Apply pagination and get results
	query = r.BuildQuery(query, pagination)
	if pagination.Sort == "" {
		query = query.Order("created_at DESC")
	}
	
	err = query.Find(&users).Error
	
	return users, err
}

func (r *userRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete user sessions first
		if err := tx.Delete(&domain.UserSession{}, "user_id = ?", id).Error; err != nil {
			return err
		}
		
		// Delete the user
		return tx.Delete(&domain.User{}, "id = ?", id).Error
	})
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID) error {
	updates := map[string]interface{}{
		"last_login_at": time.Now(),
	}
	
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepository) CreateSession(session *domain.UserSession) error {
	return r.db.Create(session).Error
}

func (r *userRepository) GetSessionByToken(token string) (*domain.UserSession, error) {
	var session domain.UserSession
	err := r.db.Preload("User").
		First(&session, "access_token = ? AND expires_at > ?", token, time.Now()).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) UpdateSession(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&domain.UserSession{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepository) DeleteSession(id uuid.UUID) error {
	return r.db.Delete(&domain.UserSession{}, "id = ?", id).Error
}

func (r *userRepository) DeleteExpiredSessions() error {
	return r.db.Delete(&domain.UserSession{}, "expires_at <= ?", time.Now()).Error
}