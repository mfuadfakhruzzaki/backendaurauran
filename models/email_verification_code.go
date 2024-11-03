// models/email_verification_token.go
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// EmailVerificationToken represents the email verification token model
type EmailVerificationToken struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    UserID    uint           `gorm:"not null;index" json:"user_id" validate:"required"`
    User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Token     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"token" validate:"required"`
    ExpiresAt time.Time      `gorm:"not null" json:"expires_at" validate:"required,gtfield=CreatedAt"`
}

// BeforeCreate GORM hook yang dijalankan sebelum membuat record baru
func (e *EmailVerificationToken) BeforeCreate(tx *gorm.DB) (err error) {
    // Validasi data token
    if err := validate.Struct(e); err != nil {
        return err
    }

    // Pastikan UserID merujuk ke user yang ada
    var user User
    if err := tx.First(&user, e.UserID).Error; err != nil {
        return fmt.Errorf("user not found: %v", err)
    }

    // Pastikan ExpiresAt lebih besar dari CreatedAt
    if !e.ExpiresAt.After(e.CreatedAt) {
        return fmt.Errorf("expires_at must be after created_at")
    }

    return nil
}
