// models/file.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// File merepresentasikan file yang diupload terkait proyek
type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID  uint           `gorm:"not null;index" json:"project_id" validate:"required"`
	Project    Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	FileName   string         `gorm:"not null" json:"file_name" validate:"required"`
	FileURL    string         `gorm:"not null" json:"file_url" validate:"required,url"`
	FileType   string         `gorm:"not null" json:"file_type" validate:"required"`
	FileSize   int64          `gorm:"not null" json:"file_size" validate:"required,gte=0"`
	UploadedBy uint           `gorm:"not null;index" json:"uploaded_by" validate:"required"`
	User       User           `gorm:"foreignKey:UploadedBy" json:"user,omitempty"`
}
