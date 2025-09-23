package core

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role defines the user roles
type Role string

const (
	AdminRole      Role = "ADMIN"
	SuperAdminRole Role = "SUPER_ADMIN"
)

// Admin represents an administrator user
type Admin struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"` // Hide password from JSON responses
	Role      Role      `gorm:"type:varchar(20);default:'ADMIN';not null" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName sets the table name for the Admin model
func (Admin) TableName() string {
	return "admins"
}

// BeforeCreate will set a UUID rather than relying on default value generation.
func (a *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}

// HashPassword hashes the admin's password using bcrypt
func (a *Admin) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	a.Password = string(bytes)
	return nil
}

// CheckPassword checks if the provided password is correct
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// LoginPayload defines the structure for the login request.
type LoginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
