// models/activity.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Activity represents an activity within a project
type Activity struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ProjectID   uint           `gorm:"not null;index" json:"project_id" validate:"required"`
	Project     Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	UserID      uint           `gorm:"not null;index" json:"user_id" validate:"required"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Description string         `gorm:"not null" json:"description" validate:"required"`
	Type        string         `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=task event milestone"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat aktivitas baru
func (a *Activity) BeforeCreate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}

// BeforeUpdate GORM hook untuk validasi sebelum memperbarui aktivitas
func (a *Activity) BeforeUpdate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}
