package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task represents a task within a project
type Task struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	ProjectID    uint           `gorm:"not null;index" json:"project_id" validate:"required"`
	Project      Project        `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	AssignedTo   *uint          `gorm:"index" json:"assigned_to,omitempty" validate:"omitempty"`
	AssignedUser User           `gorm:"foreignKey:AssignedTo" json:"assigned_user,omitempty"`
	Title        string         `gorm:"not null" json:"title" validate:"required"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	Priority     TaskPriority   `gorm:"type:varchar(10);not null" json:"priority" validate:"required,oneof=low medium high"`
	Status       TaskStatus     `gorm:"type:varchar(20);not null" json:"status" validate:"required,oneof=pending in_progress completed cancelled"`
	Deadline     *time.Time     `gorm:"type:timestamp" json:"deadline,omitempty" validate:"omitempty"`
}

// BeforeCreate GORM hook to validate before creating a new task
func (t *Task) BeforeCreate(tx *gorm.DB) (err error) {
	// Additional validation implementation if needed
	return
}

// BeforeUpdate GORM hook to validate before updating a task
func (t *Task) BeforeUpdate(tx *gorm.DB) (err error) {
	// Additional validation implementation if needed
	return
}
