// models/project.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents a project created by a user
type Project struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Title         string         `gorm:"not null" json:"title" validate:"required"`
	Description   string         `json:"description"`
	Priority      string         `json:"priority"`
	Deadline      *time.Time     `json:"deadline"`
	Status        string         `json:"status"`
	OwnerID       uint           `gorm:"not null;index" json:"owner_id" validate:"required"`
	Owner         User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Activities    []Activity     `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"activities,omitempty"`
	Notes         []Note         `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"notes,omitempty"`
	Notifications []Notification `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"notifications,omitempty"`
	Tasks         []Task         `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
	Files         []File         `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"files,omitempty"`
	Members       []User         `gorm:"many2many:project_members;constraint:OnDelete:CASCADE" json:"members,omitempty"`
	Teams         []Team         `gorm:"many2many:project_teams;constraint:OnDelete:CASCADE" json:"teams,omitempty"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat proyek baru
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}
