// models/task.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "Low"
	TaskPriorityMedium TaskPriority = "Medium"
	TaskPriorityHigh   TaskPriority = "High"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "Pending"
	TaskStatusInProgress TaskStatus = "In Progress"
	TaskStatusCompleted  TaskStatus = "Completed"
	TaskStatusCancelled  TaskStatus = "Cancelled"
)

// Task represents a task within a project
type Task struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID    uint           `gorm:"not null;index" json:"project_id" validate:"required"`
	Project      Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	AssignedToID *uint          `gorm:"index" json:"assigned_to_id,omitempty" validate:"omitempty"`
	AssignedTo   *User          `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	Title        string         `gorm:"not null" json:"title" validate:"required"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	Priority     TaskPriority   `gorm:"type:varchar(20);not null" json:"priority" validate:"required,oneof='Low' 'Medium' 'High'"`
	Status       TaskStatus     `gorm:"type:varchar(20);not null" json:"status" validate:"required,oneof='Pending' 'In Progress' 'Completed' 'Cancelled'"`
	Deadline     *time.Time     `gorm:"type:timestamp" json:"deadline,omitempty" validate:"omitempty"`
}

// BeforeCreate GORM hook to validate before creating a new task
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}

// BeforeUpdate GORM hook to validate before updating a task
func (t *Task) BeforeUpdate(tx *gorm.DB) (err error) {
	// Implementasi validasi tambahan jika diperlukan
	return
}

