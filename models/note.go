// models/note.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// NoteType merepresentasikan jenis catatan
type NoteType string

const (
    NoteTypeGeneral  NoteType = "general"
    NoteTypeActivity NoteType = "activity"
    NoteTypeProject  NoteType = "project"
)

// Note merepresentasikan catatan terkait proyek atau aktivitas
type Note struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    ProjectID uint      `gorm:"not null;index" json:"project_id" validate:"required"`
    Project   Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
    UserID    uint      `gorm:"not null;index" json:"user_id" validate:"required"`
    User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Content   string    `gorm:"not null" json:"content" validate:"required"`
    NoteType  NoteType  `gorm:"type:varchar(20);not null" json:"note_type" validate:"required,oneof=general activity project"`
}

// BeforeCreate GORM hook untuk validasi sebelum membuat catatan baru
func (n *Note) BeforeCreate(tx *gorm.DB) (err error) {
    // Implementasi validasi tambahan jika diperlukan
    return
}
