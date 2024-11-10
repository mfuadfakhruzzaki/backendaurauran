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
// CreateActivity handles the creation of a new activity within a project
func CreateActivity(c *gin.Context) {
    // Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
    userID, exists := c.Get("user_id")
    if !exists {
        utils.Logger.Warn("User ID not found in context")
        utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
        return
    }

    userIDUint, ok := userID.(uint)
    if !ok {
        utils.Logger.Warn("User ID has invalid type")
        utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
        return
    }

    utils.Logger.Infof("Creating activity for user_id: %d", userIDUint)

    // Ambil parameter project_id dari URL
    projectIDParam := c.Param("project_id")
    projectID, err := strconv.Atoi(projectIDParam)
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
        return
    }

    // Cek apakah proyek ada dan pengguna memiliki akses (pemilik atau kolaborator)
    var project models.Project
    if err := models.DB.First(&project, projectID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
            return
        }
        utils.Logger.Errorf("Failed to retrieve project: %v", err)
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
        return
    }

    // Cek akses pengguna ke proyek
    if project.OwnerID != userIDUint {
        var collaboration models.Collaboration
        if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userIDUint).First(&collaboration).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
                return
            }
            utils.Logger.Errorf("Failed to check collaboration: %v", err)
            utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
            return
        }
    }

    var req CreateActivityRequest
    // Bind JSON request ke struct
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }

    // Buat instance Activity baru dengan UserID yang benar
    activity := models.Activity{
        ProjectID:   uint(projectID),
        Description: req.Description,
        Type:        req.Type,
        UserID:      userIDUint, // Atur UserID di sini
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // Simpan activity ke database
    if err := models.DB.Create(&activity).Error; err != nil {
        utils.Logger.Errorf("Failed to create activity: %v", err)
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create activity")
        return
    }

    utils.Logger.Infof("Activity created successfully: ActivityID %d in ProjectID %d by UserID %d", activity.ID, project.ID, userIDUint)

    // Kirim respons sukses dengan data activity
    utils.CreatedResponse(c, gin.H{
        "id":          activity.ID,
        "project_id":  activity.ProjectID,
        "description": activity.Description,
        "type":        activity.Type,
        "user_id":     activity.UserID, // Sertakan UserID dalam respons
        "created_at":  activity.CreatedAt,
        "updated_at":  activity.UpdatedAt,
    })
}


// ListActivities handles retrieving all activities within a project
func ListActivities(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Cek apakah proyek ada dan pengguna memiliki akses (pemilik atau kolaborator)
	var project models.Project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek akses pengguna ke proyek
	if project.OwnerID != userID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
			return
		}
	}

	var activities []models.Activity
	// Ambil semua aktivitas dalam proyek
	if err := models.DB.Where("project_id = ?", project.ID).Find(&activities).Error; err != nil {
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
			"created_at":  activity.CreatedAt,
			"updated_at":  activity.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetActivity handles retrieving a single activity by ID within a project
func GetActivity(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan activity_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityID, err := strconv.Atoi(activityIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Cek apakah proyek ada dan pengguna memiliki akses (pemilik atau kolaborator)
	var project models.Project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek akses pengguna ke proyek
	if project.OwnerID != userID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
			return
		}
	}

	var activity models.Activity
	// Ambil aktivitas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", activityID, project.ID).First(&activity).Error; err != nil {
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
		"created_at":  activity.CreatedAt,
		"updated_at":  activity.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateActivity handles updating an activity within a project
func UpdateActivity(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan activity_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityID, err := strconv.Atoi(activityIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	var req UpdateActivityRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah setidaknya ada satu field yang diubah
	if req.Description == "" && req.Type == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Cek apakah proyek ada dan pengguna memiliki akses (pemilik atau kolaborator)
	var project models.Project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek akses pengguna ke proyek
	if project.OwnerID != userID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
			return
		}
	}

	var activity models.Activity
	// Ambil aktivitas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", activityID, project.ID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Activity not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	// Update field jika disediakan
	if req.Description != "" {
		activity.Description = req.Description
	}
	if req.Type != "" {
		activity.Type = req.Type
	}
	activity.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&activity).Error; err != nil {
		utils.Logger.Errorf("Failed to update activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	utils.Logger.Infof("Activity updated successfully: ActivityID %d in ProjectID %d", activity.ID, project.ID)

	// Siapkan data respons
	responseData := gin.H{
		"id":          activity.ID,
		"project_id":  activity.ProjectID,
		"description": activity.Description,
		"type":        activity.Type,
		"created_at":  activity.CreatedAt,
		"updated_at":  activity.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteActivity handles deleting an activity within a project
func DeleteActivity(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan activity_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	activityIDParam := c.Param("id")
	activityID, err := strconv.Atoi(activityIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Cek apakah proyek ada dan pengguna memiliki akses (pemilik atau kolaborator)
	var project models.Project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek akses pengguna ke proyek
	if project.OwnerID != userID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
			return
		}
	}

	var activity models.Activity
	// Ambil aktivitas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", activityID, project.ID).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Activity not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve activity")
		return
	}

	// Hapus aktivitas dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&activity).Error; err != nil {
		utils.Logger.Errorf("Failed to delete activity: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete activity")
		return
	}

	utils.Logger.Infof("Activity deleted successfully: ActivityID %d in ProjectID %d", activity.ID, project.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Activity deleted successfully"})
}
