// models/token.go
package models

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate *validator.Validate

// Initialize validator
func init() {
    validate = validator.New()
}

// TokenType represents the type of token
type TokenType string

const (
    TokenTypePasswordReset TokenType = "password_reset"
    TokenTypeEmailVerify   TokenType = "email_verify"
    TokenTypeJWTBlacklist  TokenType = "jwt_blacklist" // New type for blacklist
)

// Token represents the token model
type Token struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    UserID    uint           `gorm:"not null;index" json:"user_id" validate:"required"`
    User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Token     string         `gorm:"type:varchar(512);uniqueIndex;not null" json:"token" validate:"required"`
    Type      TokenType      `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=password_reset email_verify jwt_blacklist"`
    ExpiresAt time.Time      `gorm:"not null" json:"expires_at" validate:"required,gtfield=CreatedAt"`
}

// BeforeCreate GORM hook that runs before creating a new record
func (t *Token) BeforeCreate(tx *gorm.DB) (err error) {
    // Validate token data
    if err := validate.Struct(t); err != nil {
        return err
    }

    // Ensure UserID refers to an existing user
    var user User
    if err := tx.First(&user, t.UserID).Error; err != nil {
        return fmt.Errorf("user not found: %v", err)
    }

    // Ensure ExpiresAt is greater than CreatedAt
    if !t.ExpiresAt.After(t.CreatedAt) {
        return fmt.Errorf("expires_at must be after created_at")
    }

    return nil
}

// BeforeUpdate GORM hook that runs before updating a record
func (t *Token) BeforeUpdate(tx *gorm.DB) (err error) {
    // Validate token data
    if err := validate.Struct(t); err != nil {
        return err
    }

    // Ensure UserID refers to an existing user
    var user User
    if err := tx.First(&user, t.UserID).Error; err != nil {
        return fmt.Errorf("user not found: %v", err)
    }

    // Ensure ExpiresAt is greater than CreatedAt
    if !t.ExpiresAt.After(t.CreatedAt) {
        return fmt.Errorf("expires_at must be after created_at")
    }

    return nil
}
