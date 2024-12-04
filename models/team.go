// models/team.go
package models

import (
	"time"
)

type Team struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    Name        string     `gorm:"not null" json:"name"`
    Description string     `json:"description"`
    OwnerID     uint       `gorm:"not null;index" json:"owner_id"`
    Owner       User       `gorm:"foreignKey:OwnerID" json:"owner"`
    Members     []User     `gorm:"many2many:team_members;constraint:OnDelete:CASCADE" json:"members,omitempty"`
    Projects    []Project  `gorm:"many2many:project_teams;constraint:OnDelete:CASCADE" json:"projects,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `gorm:"index" json:"-"`
}
