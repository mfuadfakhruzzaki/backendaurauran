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
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the request structure for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty"`
	Description string `json:"description" binding:"omitempty"`
}

// CreateProject handles the creation of a new project
func CreateProject(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var req CreateProjectRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Buat instance Project baru
	project := models.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID.(uint),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Simpan project ke database
	if err := models.DB.Create(&project).Error; err != nil {
		utils.Logger.Errorf("Failed to create project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create project")
		return
	}

	utils.Logger.Infof("Project created successfully: %s by UserID %d", project.Name, userID.(uint))

	// Kirim respons sukses dengan data project
	utils.CreatedResponse(c, gin.H{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	})
}

// ListProjects handles retrieving all projects the user owns or collaborates on
func ListProjects(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var projects []models.Project

	// Ambil proyek yang dimiliki oleh pengguna
	if err := models.DB.Where("owner_id = ?", userID.(uint)).Find(&projects).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve owned projects: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve projects")
		return
	}

	// Tambahkan proyek yang pengguna kolaborasikan
	var collaborations []models.Collaboration
	if err := models.DB.Where("user_id = ?", userID.(uint)).Find(&collaborations).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve collaborations: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve projects")
		return
	}

	for _, collaboration := range collaborations {
		var project models.Project
		if err := models.DB.First(&project, collaboration.ProjectID).Error; err == nil {
			// Pastikan proyek tidak duplikat
			duplicate := false
			for _, p := range projects {
				if p.ID == project.ID {
					duplicate = true
					break
				}
			}
			if !duplicate {
				projects = append(projects, project)
			}
		}
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, project := range projects {
		responseData = append(responseData, gin.H{
			"id":          project.ID,
			"name":        project.Name,
			"description": project.Description,
			"owner_id":    project.OwnerID,
			"created_at":  project.CreatedAt,
			"updated_at":  project.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetProject handles retrieving a single project by ID
func GetProject(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dari URL
	projectIDParam := c.Param("project_id") // Perbaikan di sini
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var project models.Project
	// Ambil proyek dari database
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek apakah pengguna adalah pemilik atau kolaborator proyek
	if project.OwnerID != userID.(uint) {
		var collaboration models.Collaboration
		if err := models.DB.Where("project_id = ? AND user_id = ?", project.ID, userID.(uint)).First(&collaboration).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
				return
			}
			utils.Logger.Errorf("Failed to check collaboration: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
			return
		}
	}

	// Siapkan data respons
	responseData := gin.H{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateProject handles updating a project's details
func UpdateProject(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dari URL
	projectIDParam := c.Param("project_id") // Perbaikan di sini
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req UpdateProjectRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var project models.Project
	// Ambil proyek dari database
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project for update: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	// Cek apakah pengguna adalah pemilik proyek atau admin
	// Jika Anda memiliki peran admin, tambahkan logika untuk memeriksa apakah pengguna adalah admin
	// Misalnya, jika ada field role dalam user context:
	// role, exists := c.Get("role")
	// if !exists || role != "admin" {
	//     // Cek apakah pemilik proyek
	// }

	if project.OwnerID != userID.(uint) {
		// Jika Anda ingin menambahkan pengecekan peran admin, tambahkan logika di sini
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to update this project")
		return
	}

	// Update field jika disediakan
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	project.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&project).Error; err != nil {
		utils.Logger.Errorf("Failed to update project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update project")
		return
	}

	utils.Logger.Infof("Project updated successfully: ProjectID %d", project.ID)

	// Siapkan data respons
	responseData := gin.H{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"updated_at":  project.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteProject handles deleting a project
func DeleteProject(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dari URL
	projectIDParam := c.Param("project_id") // Perbaikan di sini
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var project models.Project
	// Ambil proyek dari database
	if err := models.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project for deletion: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	// Cek apakah pengguna adalah pemilik proyek atau admin
	// Jika Anda memiliki peran admin, tambahkan logika untuk memeriksa apakah pengguna adalah admin
	// Misalnya, jika ada field role dalam user context:
	// role, exists := c.Get("role")
	// if !exists || role != "admin" {
	//     // Cek apakah pemilik proyek
	// }

	if project.OwnerID != userID.(uint) {
		// Jika Anda ingin menambahkan pengecekan peran admin, tambahkan logika di sini
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this project")
		return
	}

	// Hapus proyek dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&project).Error; err != nil {
		utils.Logger.Errorf("Failed to delete project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	utils.Logger.Infof("Project deleted successfully: ProjectID %d", project.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Project deleted successfully"})
}


