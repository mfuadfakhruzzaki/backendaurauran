// controllers/project_controller.go
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

// CreateProjectRequest represents the request structure for creating a project
type CreateProjectRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `json:"status"`
	TeamIDs     []uint     `json:"team_ids"` // IDs of teams to associate with the project
}

// UpdateProjectRequest represents the request structure for updating a project
type UpdateProjectRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `json:"status"`
	TeamIDs     []uint     `json:"team_ids"` // Optional: IDs of teams to associate with the project
}

// CreateProject handles the creation of a new project
func CreateProject(c *gin.Context) {
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

	var req CreateProjectRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create a new Project instance
	project := models.Project{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Deadline:    req.Deadline,
		Status:      req.Status,
		OwnerID:     user.ID,
	}

	// Begin transaction
	tx := models.DB.Begin()
	if tx.Error != nil {
		utils.Logger.Errorf("Failed to start transaction: %v", tx.Error)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	// Save project to database
	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		utils.Logger.Errorf("Failed to create project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	// Associate teams if provided
	if len(req.TeamIDs) > 0 {
		var teams []models.Team
		if err := tx.Where("id IN ?", req.TeamIDs).Find(&teams).Error; err != nil {
			tx.Rollback()
			utils.Logger.Errorf("Failed to find teams: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to associate teams")
			return
		}

		// Associate teams with the project
		if err := tx.Model(&project).Association("Teams").Append(teams); err != nil {
			tx.Rollback()
			utils.Logger.Errorf("Failed to associate teams with project: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to associate teams")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.Logger.Errorf("Failed to commit transaction: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	utils.Logger.Infof("Project created successfully: %s by UserID %d", project.Title, user.ID)

	// Prepare response data
	responseData := gin.H{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"priority":    project.Priority,
		"deadline":    project.Deadline,
		"status":      project.Status,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
		"teams":       project.Teams,
	}

	// Send success response with project data
	utils.CreatedResponse(c, responseData)
}

// ListProjects handles retrieving all projects the user owns or collaborates on
func ListProjects(c *gin.Context) {
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

	var ownedProjects []models.Project
	// Fetch projects owned by the user
	if err := models.DB.Preload("Teams").Where("owner_id = ?", user.ID).Find(&ownedProjects).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve owned projects: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve projects")
		return
	}

	var collaboratingProjects []models.Project
	// Fetch projects associated with teams the user is a member of
	if err := models.DB.
		Joins("JOIN project_teams ON project_teams.project_id = projects.id").
		Joins("JOIN team_members ON team_members.team_id = project_teams.team_id").
		Where("team_members.user_id = ?", user.ID).
		Preload("Teams").
		Preload("Owner").
		Find(&collaboratingProjects).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve collaborating projects: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve projects")
		return
	}

	// Combine owned and collaborating projects, avoiding duplicates
	projectMap := make(map[uint]models.Project)
	for _, project := range ownedProjects {
		projectMap[project.ID] = project
	}
	for _, project := range collaboratingProjects {
		if _, exists := projectMap[project.ID]; !exists {
			projectMap[project.ID] = project
		}
	}

	// Convert map to slice
	var allProjects []models.Project
	for _, project := range projectMap {
		allProjects = append(allProjects, project)
	}

	// Prepare response data
	var responseData []gin.H
	for _, project := range allProjects {
		responseData = append(responseData, gin.H{
			"id":          project.ID,
			"title":       project.Title,
			"description": project.Description,
			"priority":    project.Priority,
			"deadline":    project.Deadline,
			"status":      project.Status,
			"owner_id":    project.OwnerID,
			"created_at":  project.CreatedAt,
			"updated_at":  project.UpdatedAt,
			"teams":       project.Teams,
		})
	}

	utils.SuccessResponse(c, responseData)
}


// GetProject handles retrieving a single project by ID
func GetProject(c *gin.Context) {
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
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var project models.Project
	// Fetch the project with preloaded Teams and Owner
	if err := models.DB.Preload("Teams").Preload("Owner").First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Check if the user is the owner
	if project.OwnerID != user.ID {
		// Check if the user is a member of any team associated with the project
		var count int64
		err = models.DB.Table("team_members").
			Where("team_id IN (SELECT team_id FROM project_teams WHERE project_id = ?) AND user_id = ?", project.ID, user.ID).
			Count(&count).Error
		if err != nil {
			utils.Logger.Errorf("Failed to check team membership: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
			return
		}
		if count == 0 && user.Role != "admin" {
			utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
			return
		}
	}

	// Prepare response data
	responseData := gin.H{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"priority":    project.Priority,
		"deadline":    project.Deadline,
		"status":      project.Status,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
		"teams":       project.Teams,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateProject handles updating a project's details
func UpdateProject(c *gin.Context) {
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
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req UpdateProjectRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var project models.Project
	// Fetch the project with preloaded Teams
	if err := models.DB.Preload("Teams").First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project for update: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	// Check if the user is the owner
	if project.OwnerID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to update this project")
		return
	}

	// Begin transaction
	tx := models.DB.Begin()
	if tx.Error != nil {
		utils.Logger.Errorf("Failed to start transaction: %v", tx.Error)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	// Update fields if provided
	if req.Title != "" {
		project.Title = req.Title
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Priority != "" {
		project.Priority = req.Priority
	}
	if req.Status != "" {
		project.Status = req.Status
	}
	if req.Deadline != nil {
		project.Deadline = req.Deadline
	}
	project.UpdatedAt = time.Now()

	// Save changes to the database
	if err := tx.Save(&project).Error; err != nil {
		tx.Rollback()
		utils.Logger.Errorf("Failed to update project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	// Update team associations if provided
	if req.TeamIDs != nil {
		var teams []models.Team
		if len(req.TeamIDs) > 0 {
			if err := tx.Where("id IN ?", req.TeamIDs).Find(&teams).Error; err != nil {
				tx.Rollback()
				utils.Logger.Errorf("Failed to find teams: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update associated teams")
				return
			}
		}
		// Replace current team associations
		if err := tx.Model(&project).Association("Teams").Replace(teams); err != nil {
			tx.Rollback()
			utils.Logger.Errorf("Failed to update associated teams: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update associated teams")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.Logger.Errorf("Failed to commit transaction: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	utils.Logger.Infof("Project updated successfully: ProjectID %d", project.ID)

	// Prepare response data
	responseData := gin.H{
		"id":          project.ID,
		"title":       project.Title,
		"description": project.Description,
		"priority":    project.Priority,
		"deadline":    project.Deadline,
		"status":      project.Status,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
		"teams":       project.Teams,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteProject handles deleting a project
func DeleteProject(c *gin.Context) {
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
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var project models.Project
	// Fetch the project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project for deletion: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	// Check if the user is the owner
	if project.OwnerID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this project")
		return
	}

	// Delete the project (soft delete if using GORM's DeletedAt)
	if err := models.DB.Delete(&project).Error; err != nil {
		utils.Logger.Errorf("Failed to delete project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	utils.Logger.Infof("Project deleted successfully: ProjectID %d", project.ID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "Project deleted successfully"})
}


