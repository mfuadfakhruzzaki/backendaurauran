// controllers/notification_controller.go
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

// NotificationController handles notification-related operations
type NotificationController struct {
	DB *gorm.DB
}

// NewNotificationController creates a new NotificationController instance
func NewNotificationController(db *gorm.DB) *NotificationController {
	return &NotificationController{
		DB: db,
	}
}

// CreateNotification handles the creation of a new notification
func (nc *NotificationController) CreateNotification(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get(utils.ContextUserKey)
	if !exists {
		utils.Logger.Warn("User not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, ok := userInterface.(models.User)
	if !ok {
		utils.Logger.Warn("User type assertion failed")
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Retrieve project_id from URL parameters
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID in URL")
		return
	}

	var req models.CreateNotificationRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warnf("Invalid request payload: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate the request using validator
	if err := utils.Validator.Struct(req); err != nil {
		utils.Logger.Warnf("Validation error: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate that the ProjectID in the body matches the URL (if provided)
	if req.ProjectID != nil && *req.ProjectID != uint(projectID) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Project ID in URL does not match Project ID in body")
		return
	}

	// Set ProjectID from URL if not provided in the body
	if req.ProjectID == nil {
		projectIDUint := uint(projectID)
		req.ProjectID = &projectIDUint
	}

	// Check if the recipient user exists
	var recipient models.User
	if err := nc.DB.First(&recipient, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Recipient user not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve recipient user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve recipient user")
		return
	}

	// Check if the project exists and if the sender has access to it
	var project models.Project
	if err := nc.DB.Preload("Owner").First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Optional: Check if the sender has access to the project (e.g., is owner or team member)
	// Implement this check based on your application's authorization logic

	// If IsRead is not provided, default to false
	isRead := false
	if req.IsRead != nil {
		isRead = *req.IsRead
	}

	// Create a new Notification instance
	notification := models.Notification{
		UserID:    req.UserID,
		Content:   req.Content,
		Type:      req.Type,
		ProjectID: req.ProjectID,
		IsRead:    isRead,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save notification to database
	if err := nc.DB.Create(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to create notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create notification")
		return
	}

	utils.Logger.Infof("Notification created successfully: NotificationID %d for UserID %d", notification.ID, req.UserID)

	// Send success response with notification data
	utils.CreatedResponse(c, gin.H{
		"id":         notification.ID,
		"user_id":    notification.UserID,
		"content":    notification.Content,
		"type":       notification.Type,
		"project_id": notification.ProjectID,
		"is_read":    notification.IsRead,
		"created_at": notification.CreatedAt,
		"updated_at": notification.UpdatedAt,
	})
}

// ListNotifications handles retrieving all notifications for the current user
func (nc *NotificationController) ListNotifications(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get(utils.ContextUserKey)
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

	var notifications []models.Notification
	// Retrieve all notifications for the user, ordered by newest first
	if err := nc.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&notifications).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve notifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notifications")
		return
	}

	// Prepare response data
	var responseData []gin.H
	for _, notification := range notifications {
		responseData = append(responseData, gin.H{
			"id":         notification.ID,
			"user_id":    notification.UserID,
			"content":    notification.Content,
			"type":       notification.Type,
			"project_id": notification.ProjectID,
			"is_read":    notification.IsRead,
			"created_at": notification.CreatedAt,
			"updated_at": notification.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetNotification handles retrieving a single notification by ID for the current user
func (nc *NotificationController) GetNotification(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get(utils.ContextUserKey)
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

	// Retrieve notification_id from URL parameters
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var notification models.Notification
	// Retrieve notification from database
	if err := nc.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Prepare response data
	responseData := gin.H{
		"id":         notification.ID,
		"user_id":    notification.UserID,
		"content":    notification.Content,
		"type":       notification.Type,
		"project_id": notification.ProjectID,
		"is_read":    notification.IsRead,
		"created_at": notification.CreatedAt,
		"updated_at": notification.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateNotification handles updating a notification's status or content
func (nc *NotificationController) UpdateNotification(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get(utils.ContextUserKey)
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

	// Retrieve notification_id from URL parameters
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var req models.UpdateNotificationRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request using validator
	if err := utils.Validator.Struct(req); err != nil {
		utils.Logger.Warnf("Validation error: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if at least one field is provided for update
	if req.Content == nil && req.IsRead == nil && req.Type == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	var notification models.Notification
	// Retrieve notification from database
	if err := nc.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Update fields if provided
	if req.Content != nil {
		notification.Content = *req.Content
	}
	if req.Type != nil {
		notification.Type = models.NotificationType(*req.Type)
	}
	if req.IsRead != nil {
		notification.IsRead = *req.IsRead
	}
	notification.UpdatedAt = time.Now()

	// Save changes to database
	if err := nc.DB.Save(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to update notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification")
		return
	}

	utils.Logger.Infof("Notification updated successfully: NotificationID %d for UserID %d", notification.ID, notification.UserID)

	// Prepare response data
	responseData := gin.H{
		"id":         notification.ID,
		"user_id":    notification.UserID,
		"content":    notification.Content,
		"type":       notification.Type,
		"project_id": notification.ProjectID,
		"is_read":    notification.IsRead,
		"created_at": notification.CreatedAt,
		"updated_at": notification.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteNotification handles deleting a notification
func (nc *NotificationController) DeleteNotification(c *gin.Context) {
	// Retrieve the User object from context set by AuthMiddleware
	userInterface, exists := c.Get(utils.ContextUserKey)
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

	// Retrieve notification_id from URL parameters
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var notification models.Notification
	// Retrieve notification from database
	if err := nc.DB.Where("id = ? AND user_id = ?", notificationID, user.ID).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Delete notification from database (soft delete if using GORM's DeletedAt)
	if err := nc.DB.Delete(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to delete notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	utils.Logger.Infof("Notification deleted successfully: NotificationID %d for UserID %d", notification.ID, notification.UserID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "Notification deleted successfully"})
}
