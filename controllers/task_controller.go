// controllers/task_controller.go
package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// CreateTaskRequest represents the request structure for creating a new task
type CreateTaskRequest struct {
	Title        string              `json:"title" binding:"required"`
	Description  string              `json:"description" binding:"required"`
	Priority     models.TaskPriority `json:"priority" binding:"required,oneof='Low' 'Medium' 'High'"`
	Status       models.TaskStatus   `json:"status" binding:"required,oneof='Pending' 'In Progress' 'Completed' 'Cancelled'"`
	Deadline     *time.Time          `json:"deadline" binding:"omitempty"`
	AssignedToID *uint               `json:"assigned_to_id" binding:"omitempty"`
}

// UpdateTaskRequest represents the request structure for updating a task
type UpdateTaskRequest struct {
	Title        *string              `json:"title" binding:"omitempty"`
	Description  *string              `json:"description" binding:"omitempty"`
	Priority     *models.TaskPriority `json:"priority" binding:"omitempty,oneof='Low' 'Medium' 'High'"`
	Status       *models.TaskStatus   `json:"status" binding:"omitempty,oneof='Pending' 'In Progress' 'Completed' 'Cancelled'"`
	Deadline     *time.Time           `json:"deadline" binding:"omitempty"`
	AssignedToID *uint                `json:"assigned_to_id" binding:"omitempty"`
}

// CreateTask handles the creation of a new task within a specific project
func CreateTask(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the project exists and the user has access
	var project models.Project
	if err := models.DB.Preload("Teams.Members").First(&project, uint(projectID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create task")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var req CreateTaskRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// If AssignedToID is provided, check if the assigned user exists and has access
	if req.AssignedToID != nil {
		var assignedUser models.User
		if err := models.DB.First(&assignedUser, *req.AssignedToID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusNotFound, "Assigned user not found")
				return
			}
			utils.Logger.Errorf("Failed to retrieve assigned user: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve assigned user")
			return
		}
		// Check if the assigned user is a member of any team associated with the project
		isMember, err := models.UserIsMemberOfProjectTeams(*req.AssignedToID, uint(projectID))
		if err != nil {
			utils.Logger.Errorf("Failed to check if user is member of project teams: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create task")
			return
		}
		if !isMember {
			utils.ErrorResponse(c, http.StatusBadRequest, "Assigned user is not a member of the project's teams")
			return
		}
	}

	// Create a new Task instance
	task := models.Task{
		ProjectID:    uint(projectID),
		AssignedToID: req.AssignedToID,
		Title:        req.Title,
		Description:  req.Description,
		Priority:     req.Priority,
		Status:       req.Status,
		Deadline:     req.Deadline,
	}

	// Save task to database
	if err := models.DB.Create(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to create task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create task")
		return
	}

	utils.Logger.Infof("Task created successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, user.ID)

	// Prepare response data
	responseData := gin.H{
		"id":           task.ID,
		"title":        task.Title,
		"description":  task.Description,
		"priority":     task.Priority,
		"status":       task.Status,
		"deadline":     task.Deadline,
		"assigned_to":  task.AssignedToID,
		"project_id":   task.ProjectID,
		"created_at":   task.CreatedAt,
		"updated_at":   task.UpdatedAt,
	}

	// Send success response with task data
	utils.CreatedResponse(c, responseData)
}

// ListTasks handles retrieving all tasks within a specific project
func ListTasks(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var tasks []models.Task
	// Retrieve all tasks for the project, including AssignedTo user
	if err := models.DB.Where("project_id = ?", uint(projectID)).
		Preload("AssignedTo").
		Order("created_at desc").
		Find(&tasks).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve tasks: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}

	// Prepare response data
	var responseData []gin.H
	for _, task := range tasks {
		responseData = append(responseData, gin.H{
			"id":            task.ID,
			"title":         task.Title,
			"description":   task.Description,
			"priority":      task.Priority,
			"status":        task.Status,
			"deadline":      task.Deadline,
			"assigned_to_id": task.AssignedToID,
			"project_id":    task.ProjectID,
			"created_at":    task.CreatedAt,
			"updated_at":    task.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetTask handles retrieving a specific task within a project
func GetTask(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id and task_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var task models.Task
	// Retrieve the task from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(taskID), uint(projectID)).
		Preload("AssignedTo").
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// Prepare response data
	responseData := gin.H{
		"id":            task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"priority":      task.Priority,
		"status":        task.Status,
		"deadline":      task.Deadline,
		"assigned_to_id": task.AssignedToID,
		"project_id":    task.ProjectID,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateTask handles updating a specific task within a project
func UpdateTask(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id and task_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req UpdateTaskRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if at least one field is provided for update
	if req.Title == nil && req.Description == nil && req.Priority == nil && req.Status == nil && req.Deadline == nil && req.AssignedToID == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update task")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var task models.Task
	// Retrieve the task from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(taskID), uint(projectID)).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// If AssignedToID is provided, check if the assigned user exists and has access
	if req.AssignedToID != nil {
		if *req.AssignedToID != 0 {
			var assignedUser models.User
			if err := models.DB.First(&assignedUser, *req.AssignedToID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					utils.ErrorResponse(c, http.StatusNotFound, "Assigned user not found")
					return
				}
				utils.Logger.Errorf("Failed to retrieve assigned user: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve assigned user")
				return
			}
			// Check if the assigned user is a member of any team associated with the project
			isMember, err := models.UserIsMemberOfProjectTeams(*req.AssignedToID, uint(projectID))
			if err != nil {
				utils.Logger.Errorf("Failed to check if user is member of project teams: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update task")
				return
			}
			if !isMember {
				utils.ErrorResponse(c, http.StatusBadRequest, "Assigned user is not a member of the project's teams")
				return
			}
			task.AssignedToID = req.AssignedToID
		} else {
			// If AssignedToID is 0, remove the assignment
			task.AssignedToID = nil
		}
	}

	// Update fields if provided
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}
	task.UpdatedAt = time.Now()

	// Save changes to the database
	if err := models.DB.Save(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to update task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update task")
		return
	}

	utils.Logger.Infof("Task updated successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, user.ID)

	// Retrieve the updated task with AssignedTo user
	if err := models.DB.Preload("AssignedTo").First(&task, task.ID).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve updated task: %v", err)
		// Meskipun gagal mengambil task yang diperbarui, tetap kirim respons sukses
	}

	// Prepare response data
	responseData := gin.H{
		"id":            task.ID,
		"title":         task.Title,
		"description":   task.Description,
		"priority":      task.Priority,
		"status":        task.Status,
		"deadline":      task.Deadline,
		"assigned_to_id": task.AssignedToID,
		"project_id":    task.ProjectID,
		"created_at":    task.CreatedAt,
		"updated_at":    task.UpdatedAt,
	}

	// Send success response
	utils.SuccessResponse(c, responseData)
}

// DeleteTask handles deleting a specific task within a project
func DeleteTask(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id and task_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("task_id")
	taskID, err := strconv.ParseUint(taskIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var task models.Task
	// Retrieve the task from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(taskID), uint(projectID)).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// Delete the task from the database (soft delete if using gorm.DeletedAt)
	if err := models.DB.Delete(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to delete task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	utils.Logger.Infof("Task deleted successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, user.ID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "Task deleted successfully"})
}
