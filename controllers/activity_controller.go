// controllers/activity_controller.go
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

// CreateActivityRequest represents the request structure for creating an activity
type CreateActivityRequest struct {
	Description string `json:"description" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=task event milestone"`
}

// UpdateActivityRequest represents the request structure for updating an activity
type UpdateActivityRequest struct {
	Description string `json:"description" binding:"omitempty"`
	Type        string `json:"type" binding:"omitempty,oneof=task event milestone"`
}

// CreateActivity handles the creation of a new activity within a project
func CreateActivity(c *gin.Context) {
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

	utils.Logger.Infof("Creating activity for user_id: %d", user.ID)

	// Retrieve project_id from URL parameters and convert to uint
	projectIDParam := c.Param("project_id")
	projectIDUint, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectIDUint))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create activity")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var req CreateActivityRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Tambahkan logging untuk memeriksa nilai Type yang diterima
	utils.Logger.Infof("Received Type: %s", req.Type)

	// Konversi Type dari string ke models.ActivityType
	activityType := models.ActivityType(req.Type)

	// Buat instance Activity baru
	activity := models.Activity{
		ProjectID:   uint(projectIDUint),
		Description: req.Description,
		Type:        activityType,
		UserID:      user.ID, // Set UserID di sini
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Simpan activity ke database
	if err := models.DB.Create(&activity).Error; err != nil {
		utils.Logger.Errorf("Failed to create activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create activity")
		return
	}

	utils.Logger.Infof("Activity created successfully: ActivityID %d in ProjectID %d by UserID %d", activity.ID, projectIDUint, user.ID)

	// Siapkan data respons
	responseData := gin.H{
		"id":          activity.ID,
		"project_id":  activity.ProjectID,
		"description": activity.Description,
		"type":        activity.Type,
		"user_id":     activity.UserID, // Sertakan UserID dalam respons
		"created_at":  activity.CreatedAt,
		"updated_at":  activity.UpdatedAt,
	}

	// Kirim respons sukses dengan data activity
	utils.CreatedResponse(c, responseData)
}

// ListActivities handles retrieving all activities within a project
func ListActivities(c *gin.Context) {
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

	// Retrieve project_id dari URL parameters dan konversi ke uint
	projectIDParam := c.Param("project_id")
	projectIDUint, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectIDUint))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activities")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var activities []models.Activity
	// Retrieve all activities for the project, termasuk User yang membuat setiap activity
	if err := models.DB.Where("project_id = ?", uint(projectIDUint)).
		Preload("User").
		Order("created_at desc").
		Find(&activities).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve activities: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activities")
		return
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, activity := range activities {
		responseData = append(responseData, gin.H{
			"id":          activity.ID,
			"project_id":  activity.ProjectID,
			"description": activity.Description,
			"type":        activity.Type,
			"user_id":     activity.UserID,
			"created_at":  activity.CreatedAt,
			"updated_at":  activity.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetActivity handles retrieving a single activity by ID within a project
func GetActivity(c *gin.Context) {
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

	// Retrieve project_id dan activity_id dari URL parameters dan konversi ke uint
	projectIDParam := c.Param("project_id")
	projectIDUint, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityIDUint, err := strconv.ParseUint(activityIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectIDUint))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var activity models.Activity
	// Retrieve the activity from the database, termasuk User yang membuatnya
	if err := models.DB.Where("id = ? AND project_id = ?", uint(activityIDUint), uint(projectIDUint)).
		Preload("User").
		First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Activity not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	// Siapkan data respons
	responseData := gin.H{
		"id":          activity.ID,
		"project_id":  activity.ProjectID,
		"description": activity.Description,
		"type":        activity.Type,
		"user_id":     activity.UserID,
		"created_at":  activity.CreatedAt,
		"updated_at":  activity.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateActivity handles updating an activity within a project
func UpdateActivity(c *gin.Context) {
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

	// Retrieve project_id dan activity_id dari URL parameters dan konversi ke uint
	projectIDParam := c.Param("project_id")
	projectIDUint, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityIDUint, err := strconv.ParseUint(activityIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	var req UpdateActivityRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if at least one field is provided for update
	if req.Description == "" && req.Type == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectIDUint))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update activity")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var activity models.Activity
	// Retrieve the activity dari database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(activityIDUint), uint(projectIDUint)).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Activity not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	// Update fields jika disediakan
	if req.Description != "" {
		activity.Description = req.Description
	}
	if req.Type != "" {
		activity.Type = models.ActivityType(req.Type)
	}
	activity.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&activity).Error; err != nil {
		utils.Logger.Errorf("Failed to update activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	utils.Logger.Infof("Activity updated successfully: ActivityID %d in ProjectID %d by UserID %d", activity.ID, projectIDUint, user.ID)

	// Siapkan data respons
	responseData := gin.H{
		"id":          activity.ID,
		"project_id":  activity.ProjectID,
		"description": activity.Description,
		"type":        activity.Type,
		"user_id":     activity.UserID,
		"created_at":  activity.CreatedAt,
		"updated_at":  activity.UpdatedAt,
	}

	// Kirim respons sukses
	utils.SuccessResponse(c, responseData)
}

// DeleteActivity handles deleting an activity within a project
func DeleteActivity(c *gin.Context) {
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

	// Retrieve project_id dan activity_id dari URL parameters dan konversi ke uint
	projectIDParam := c.Param("project_id")
	projectIDUint, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityIDUint, err := strconv.ParseUint(activityIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectIDUint))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete activity")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var activity models.Activity
	// Retrieve the activity dari database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(activityIDUint), uint(projectIDUint)).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Activity not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	// Delete the activity dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&activity).Error; err != nil {
		utils.Logger.Errorf("Failed to delete activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete activity")
		return
	}

	utils.Logger.Infof("Activity deleted successfully: ActivityID %d in ProjectID %d by UserID %d", activity.ID, projectIDUint, user.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Activity deleted successfully"})
}
