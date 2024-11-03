// models/collaboration.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// CollaborationRole merepresentasikan peran kolaborator dalam proyek
type CollaborationRole string

const (
	CollaborationRoleAdmin        CollaborationRole = "admin"
	CollaborationRoleCollaborator CollaborationRole = "collaborator"
)

// Collaboration merepresentasikan kolaborasi pengguna dalam proyek
type Collaboration struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	ProjectID uint              `gorm:"not null;index" json:"project_id" validate:"required"`
	Project   Project           `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	UserID    uint              `gorm:"not null;index" json:"user_id" validate:"required"`
	User      User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role      CollaborationRole `gorm:"type:varchar(20);not null" json:"role" validate:"required,oneof=admin collaborator"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat kolaborasi baru
func (c *Collaboration) BeforeCreate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}

// BeforeUpdate GORM hook untuk validasi sebelum memperbarui kolaborasi
func (c *Collaboration) BeforeUpdate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}
