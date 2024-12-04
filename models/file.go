// models/file.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// File represents a file within a project
type File struct {
    ID         uint           `gorm:"primaryKey" json:"id"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    ProjectID  uint           `gorm:"not null;index" json:"project_id" validate:"required"`
    Project    Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    UploadedBy uint           `gorm:"not null;index" json:"uploaded_by" validate:"required"`
    User       User           `gorm:"foreignKey:UploadedBy" json:"user,omitempty"`
    Filename   string         `gorm:"not null;column:filename" json:"filename" validate:"required"` // Sesuaikan tag kolom
    FileURL    string         `gorm:"not null" json:"file_url" validate:"required,url"`
    FileType   string         `gorm:"type:varchar(50);not null" json:"file_type" validate:"required,oneof=image/jpeg image/png image/gif application/pdf video/mp4"`
    FileSize   int64          `gorm:"not null" json:"file_size" validate:"required,gte=0"`
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate GORM hook for validation before creating a file
func (f *File) BeforeCreate(tx *gorm.DB) (err error) {
    // Implement additional validation if necessary
    return
}

// BeforeUpdate GORM hook for validation before updating a file
func (f *File) BeforeUpdate(tx *gorm.DB) (err error) {
    // Implement additional validation if necessary
    return
}
