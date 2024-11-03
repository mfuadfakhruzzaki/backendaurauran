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

// CreateNotificationRequest represents the request structure for creating a notification
type CreateNotificationRequest struct {
	UserID    uint                 `json:"user_id" binding:"required"`
	Content   string               `json:"content" binding:"required"`
	Type      models.NotificationType `json:"type" binding:"required,oneof=info warning error success"`
	ProjectID *uint                `json:"project_id" binding:"omitempty"`
	IsRead    *bool                `json:"is_read"` // Opsional, default false
}

// UpdateNotificationRequest represents the request structure for updating a notification
type UpdateNotificationRequest struct {
	Content *string                 `json:"content" binding:"omitempty"`
	Type    *models.NotificationType `json:"type" binding:"omitempty,oneof=info warning error success"`
	IsRead  *bool                   `json:"is_read" binding:"omitempty"`
}

// CreateNotification handles the creation of a new notification
func CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah pengguna yang dituju ada
	var user models.User
	if err := models.DB.First(&user, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	// Jika ProjectID disediakan, cek apakah proyek ada
	if req.ProjectID != nil {
		var project models.Project
		if err := models.DB.First(&project, *req.ProjectID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
				return
			}
			utils.Logger.Errorf("Failed to retrieve project: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
			return
		}
	}

	// Jika IsRead tidak disediakan, set default false
	isRead := false
	if req.IsRead != nil {
		isRead = *req.IsRead
	}

	// Buat instance Notification baru
	notification := models.Notification{
		UserID:    req.UserID,
		Content:   req.Content,
		Type:      req.Type,
		IsRead:    isRead,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Jika ProjectID disediakan, set
	if req.ProjectID != nil {
		notification.ProjectID = *req.ProjectID
	}

	// Simpan notification ke database
	if err := models.DB.Create(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to create notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create notification")
		return
	}

	utils.Logger.Infof("Notification created successfully: NotificationID %d for UserID %d", notification.ID, req.UserID)

	// Kirim respons sukses dengan data notifikasi
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
func ListNotifications(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var notifications []models.Notification
	// Ambil semua notifikasi untuk pengguna, urutkan dari yang terbaru
	if err := models.DB.Where("user_id = ?", currentUserID.(uint)).Order("created_at desc").Find(&notifications).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve notifications: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notifications")
		return
	}

	// Siapkan data respons
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
func GetNotification(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter notification_id dari URL
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var notification models.Notification
	// Ambil notifikasi dari database
	if err := models.DB.Where("id = ? AND user_id = ?", notificationID, currentUserID.(uint)).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Siapkan data respons
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
func UpdateNotification(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter notification_id dari URL
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var req UpdateNotificationRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah ada setidaknya satu field yang diubah
	if req.Content == nil && req.IsRead == nil && req.Type == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	var notification models.Notification
	// Ambil notifikasi dari database
	if err := models.DB.Where("id = ? AND user_id = ?", notificationID, currentUserID.(uint)).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Update field jika disediakan
	if req.Content != nil {
		notification.Content = *req.Content
	}
	if req.Type != nil {
		notification.Type = *req.Type
	}
	if req.IsRead != nil {
		notification.IsRead = *req.IsRead
	}
	notification.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to update notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update notification")
		return
	}

	utils.Logger.Infof("Notification updated successfully: NotificationID %d for UserID %d", notification.ID, notification.UserID)

	// Siapkan data respons
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
func DeleteNotification(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter notification_id dari URL
	notificationIDParam := c.Param("id")
	notificationID, err := strconv.Atoi(notificationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	var notification models.Notification
	// Ambil notifikasi dari database
	if err := models.DB.Where("id = ? AND user_id = ?", notificationID, currentUserID.(uint)).First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Notification not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notification")
		return
	}

	// Hapus notifikasi dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&notification).Error; err != nil {
		utils.Logger.Errorf("Failed to delete notification: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete notification")
		return
	}

	utils.Logger.Infof("Notification deleted successfully: NotificationID %d for UserID %d", notification.ID, notification.UserID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Notification deleted successfully"})
}
