package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleManager   UserRole = "manager"
	UserRoleViewer    UserRole = "viewer"
	UserRoleTechnician UserRole = "technician"
)

// User represents an asset manager or system user
// We keep this simple for MVP but extensible for enterprise features
type User struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string         `json:"email" gorm:"type:varchar(255);unique;not null"`
	PasswordHash  string         `json:"-" gorm:"type:varchar(255);not null"` // Never expose in JSON
	FullName      string         `json:"full_name" gorm:"type:varchar(255)"`
	AvatarURL     string         `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
	Settings      JSON           `json:"settings" gorm:"type:jsonb;default:'{}'"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// SetPassword hashes and sets the user's password
// Using bcrypt for secure password hashing
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password against the hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// RefreshToken stores refresh tokens with metadata
// This allows us to revoke specific tokens and track devices
type RefreshToken struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Token      string    `json:"-" gorm:"type:varchar(255);unique;not null"`
	DeviceInfo string    `json:"device_info,omitempty" gorm:"type:varchar(500)"`
	IPAddress  string    `json:"ip_address,omitempty" gorm:"type:varchar(45)"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserSession represents an active user session
type UserSession struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	AccessToken  string    `json:"-" gorm:"type:varchar(255);unique;not null"`
	RefreshToken string    `json:"-" gorm:"type:varchar(255);unique;not null"`
	DeviceInfo   string    `json:"device_info,omitempty" gorm:"type:varchar(500)"`
	IPAddress    string    `json:"ip_address,omitempty" gorm:"type:varchar(45)"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	
	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// LoginRequest for authentication endpoint
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest for user registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=12,password"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
}

// AuthResponse returned after successful login
type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"` // Only in cookie, not in JSON
}

// TokenClaims for JWT tokens
type TokenClaims struct {
	UserID    uuid.UUID `json:"sub"`
	Email     string    `json:"email"`
	TokenType string    `json:"typ"` // "access" or "refresh"
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
}