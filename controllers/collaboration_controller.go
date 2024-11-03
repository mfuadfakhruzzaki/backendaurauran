// controllers/collaboration_controller.go
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

// AddCollaboratorRequest represents the request structure for adding a collaborator
type AddCollaboratorRequest struct {
    UserID uint                     `json:"user_id" binding:"required"`
    Role   models.CollaborationRole `json:"role" binding:"required,oneof=admin collaborator"`
}

type UpdateCollaboratorRoleRequest struct {
    Role models.CollaborationRole `json:"role" binding:"required,oneof=admin collaborator"`
}

// AddCollaborator handles adding a new collaborator to a project
func AddCollaborator(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
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

	// Bind JSON request ke struct
	var req AddCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah proyek ada
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

	// Cek apakah pengguna saat ini adalah pemilik proyek atau admin
	if project.OwnerID != currentUserID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to add collaborators to this project")
		return
	}

	// Cek apakah pengguna yang akan ditambahkan ada
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

	// Cek apakah pengguna sudah menjadi kolaborator
	var existingCollab models.Collaboration
	if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, req.UserID).First(&existingCollab).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "User is already a collaborator")
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.Logger.Errorf("Failed to check existing collaboration: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add collaborator")
		return
	}

	// Buat kolaborasi baru
	collaboration := models.Collaboration{
		ProjectID: project.ID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	// Simpan kolaborasi ke database
	if err := models.DB.Create(&collaboration).Error; err != nil {
		utils.Logger.Errorf("Failed to add collaborator: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add collaborator")
		return
	}

	utils.Logger.Infof("Collaborator added successfully: UserID %d to ProjectID %d", req.UserID, project.ID)

	// Kirim respons sukses dengan data kolaborasi
	utils.CreatedResponse(c, gin.H{
		"id":         collaboration.ID,
		"project_id": collaboration.ProjectID,
		"user_id":    collaboration.UserID,
		"role":       collaboration.Role,
		"created_at": collaboration.CreatedAt,
		"updated_at": collaboration.UpdatedAt,
	})
}

// RemoveCollaborator handles removing a collaborator from a project
func RemoveCollaborator(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan collaborator_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	collaboratorIDParam := c.Param("collaborator_id")
	collaboratorID, err := strconv.Atoi(collaboratorIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid collaborator ID")
		return
	}

	// Cek apakah proyek ada
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

	// Cek apakah pengguna saat ini adalah pemilik proyek atau admin
	if project.OwnerID != currentUserID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to remove collaborators from this project")
		return
	}

	// Cek apakah kolaborator ada
	var collaboration models.Collaboration
	if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, collaboratorID).First(&collaboration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Collaborator not found in this project")
			return
		}
		utils.Logger.Errorf("Failed to retrieve collaboration: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove collaborator")
		return
	}

	// Hapus kolaborasi dari database
	if err := models.DB.Delete(&collaboration).Error; err != nil {
		utils.Logger.Errorf("Failed to remove collaborator: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove collaborator")
		return
	}

	utils.Logger.Infof("Collaborator removed successfully: UserID %d from ProjectID %d", collaboratorID, project.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Collaborator removed successfully"})
}

// UpdateCollaboratorRole handles updating a collaborator's role within a project
func UpdateCollaboratorRole(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan collaborator_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	collaboratorIDParam := c.Param("collaborator_id")
	collaboratorID, err := strconv.Atoi(collaboratorIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid collaborator ID")
		return
	}

	// Bind JSON request ke struct
	var req UpdateCollaboratorRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah proyek ada
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

	// Cek apakah pengguna saat ini adalah pemilik proyek atau admin
	if project.OwnerID != currentUserID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to update collaborators in this project")
		return
	}

	// Cek apakah kolaborator ada
	var collaboration models.Collaboration
	if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, collaboratorID).First(&collaboration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Collaborator not found in this project")
			return
		}
		utils.Logger.Errorf("Failed to retrieve collaboration: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update collaborator")
		return
	}

	// Update peran kolaborator
	collaboration.Role = req.Role
	collaboration.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&collaboration).Error; err != nil {
		utils.Logger.Errorf("Failed to update collaborator role: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update collaborator role")
		return
	}

	utils.Logger.Infof("Collaborator role updated successfully: UserID %d in ProjectID %d to role %s", collaboratorID, project.ID, req.Role)

	// Kirim respons sukses dengan data kolaborasi yang diperbarui
	utils.SuccessResponse(c, gin.H{
		"id":         collaboration.ID,
		"project_id": collaboration.ProjectID,
		"user_id":    collaboration.UserID,
		"role":       collaboration.Role,
		"updated_at": collaboration.UpdatedAt,
	})
}

// ListCollaborators handles retrieving all collaborators of a project
func ListCollaborators(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
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

	// Cek apakah proyek ada
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

	// Cek apakah pengguna saat ini adalah pemilik proyek atau kolaborator
	if project.OwnerID != currentUserID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, currentUserID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to check access permissions")
			return
		}
	}

	// Ambil semua kolaborator dalam proyek
	var collaborations []models.Collaboration
	if err := models.DB.Preload("User").Where("project_id = ?", project.ID).Find(&collaborations).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve collaborators: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve collaborators")
		return
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, collab := range collaborations {
		responseData = append(responseData, gin.H{
			"id":         collab.ID,
			"user_id":    collab.UserID,
			"user_email": collab.User.Email,
			"role":       collab.Role,
			"created_at": collab.CreatedAt,
			"updated_at": collab.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}
