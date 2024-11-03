// models/project.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a project created by a user
type Project struct {
    ID               uint           `gorm:"primaryKey" json:"id"`
    CreatedAt        time.Time      `json:"created_at"`
    UpdatedAt        time.Time      `json:"updated_at"`
    Name             string         `gorm:"not null" json:"name" validate:"required"`
    Description      string         `json:"description"`
    OwnerID          uint           `gorm:"not null;index" json:"owner_id" validate:"required"`
    Owner            User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
    Collaborations   []Collaboration `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"collaborations,omitempty"`
    Activities       []Activity     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"activities,omitempty"`
    Notes            []Note         `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"notes,omitempty"`
    Notifications    []Notification `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"notifications,omitempty"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat proyek baru
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
    // Implementasi validasi tambahan jika diperlukan
    return
}
