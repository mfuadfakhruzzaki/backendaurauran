// models/notification.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType merepresentasikan jenis notifikasi
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

// Notification merepresentasikan notifikasi yang dikirim ke pengguna
type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
	ProjectID uint             `gorm:"index" json:"project_id,omitempty" validate:"omitempty"`
	Project   Project          `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	UserID    uint             `gorm:"not null;index" json:"user_id" validate:"required"`
	User      User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string           `gorm:"not null" json:"content" validate:"required"`
	Type      NotificationType `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=info warning error success"`
	IsRead    bool             `gorm:"default:false" json:"is_read"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat notifikasi baru
func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}

// BeforeUpdate GORM hook untuk validasi sebelum memperbarui notifikasi
func (n *Notification) BeforeUpdate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}
