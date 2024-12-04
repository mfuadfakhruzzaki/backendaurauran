// models/notification.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

// Notification represents a notification sent to a user.
type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
	ProjectID *uint            `gorm:"index" json:"project_id,omitempty" validate:"omitempty"` // Made nullable
	Project   *Project         `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	UserID    uint             `gorm:"not null;index" json:"user_id" validate:"required"`
	User      *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string           `gorm:"not null" json:"content" validate:"required"`
	Type      NotificationType `gorm:"type:varchar(20);not null" json:"type" validate:"required,oneof=info warning error success"`
	IsRead    bool             `gorm:"default:false" json:"is_read"`
}

// CreateNotificationRequest represents the request payload for creating a notification
type CreateNotificationRequest struct {
	UserID    uint               `json:"user_id" binding:"required"` // Recipient User ID
	Content   string             `json:"content" binding:"required"`
	Type      NotificationType   `json:"type" binding:"required,oneof=info warning error success"`
	ProjectID *uint              `json:"project_id" binding:"omitempty"`
	IsRead    *bool              `json:"is_read"` // Optional, default false
}

// UpdateNotificationRequest represents the request payload for updating a notification

type UpdateNotificationRequest struct {

    Content *string `json:"content,omitempty"`

    Type    *string `json:"type,omitempty"`
	
    IsRead  *bool   `json:"is_read,omitempty"`

}

// BeforeCreate GORM hook for additional validation before creating a notification.
func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	// Implement additional validation if needed.
	// For example, ensure that the Type is valid, or that ProjectID references an existing project.
	return
}

// BeforeUpdate GORM hook for additional validation before updating a notification.
func (n *Notification) BeforeUpdate(tx *gorm.DB) (err error) {
	// Implement additional validation if needed.
	// For example, prevent changing the UserID or ensure that updates are permissible.
	return
}
