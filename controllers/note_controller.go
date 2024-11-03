// controllers/note_controller.go
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

// CreateNoteRequest represents the request structure for creating a note
type CreateNoteRequest struct {
	Content string `json:"content" binding:"required"`
}

// UpdateNoteRequest represents the request structure for updating a note
type UpdateNoteRequest struct {
	Content string `json:"content" binding:"omitempty"`
}

// CreateNote handles the creation of a new note within a project
func CreateNote(c *gin.Context) {
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

	var req CreateNoteRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Buat instance Note baru
	note := models.Note{
		ProjectID: project.ID,
		UserID:    userID.(uint),
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Simpan note ke database
	if err := models.DB.Create(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to create note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create note")
		return
	}

	utils.Logger.Infof("Note created successfully: NoteID %d in ProjectID %d", note.ID, project.ID)

	// Kirim respons sukses dengan data note
	utils.CreatedResponse(c, gin.H{
		"id":         note.ID,
		"project_id": note.ProjectID,
		"user_id":    note.UserID,
		"content":    note.Content,
		"created_at": note.CreatedAt,
		"updated_at": note.UpdatedAt,
	})
}

// ListNotes handles retrieving all notes within a project
func ListNotes(c *gin.Context) {
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

	var notes []models.Note
	// Ambil semua catatan dalam proyek
	if err := models.DB.Where("project_id = ?", project.ID).Find(&notes).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve notes: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, note := range notes {
		responseData = append(responseData, gin.H{
			"id":         note.ID,
			"project_id": note.ProjectID,
			"user_id":    note.UserID,
			"content":    note.Content,
			"created_at": note.CreatedAt,
			"updated_at": note.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetNote handles retrieving a single note by ID within a project
func GetNote(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan note_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	noteIDParam := c.Param("id")
	noteID, err := strconv.Atoi(noteIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid note ID")
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

	var note models.Note
	// Ambil catatan dari database
	if err := models.DB.Where("id = ? AND project_id = ?", noteID, project.ID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Siapkan data respons
	responseData := gin.H{
		"id":         note.ID,
		"project_id": note.ProjectID,
		"user_id":    note.UserID,
		"content":    note.Content,
		"created_at": note.CreatedAt,
		"updated_at": note.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateNote handles updating a note within a project
func UpdateNote(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan note_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	noteIDParam := c.Param("id")
	noteID, err := strconv.Atoi(noteIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var req UpdateNoteRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah setidaknya ada satu field yang diubah
	if req.Content == "" {
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

	var note models.Note
	// Ambil catatan dari database
	if err := models.DB.Where("id = ? AND project_id = ?", noteID, project.ID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Update field jika disediakan
	if req.Content != "" {
		note.Content = req.Content
	}
	note.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to update note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update note")
		return
	}

	utils.Logger.Infof("Note updated successfully: NoteID %d in ProjectID %d", note.ID, project.ID)

	// Siapkan data respons
	responseData := gin.H{
		"id":         note.ID,
		"project_id": note.ProjectID,
		"user_id":    note.UserID,
		"content":    note.Content,
		"created_at": note.CreatedAt,
		"updated_at": note.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteNote handles deleting a note within a project
func DeleteNote(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Ambil parameter project_id dan note_id dari URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	noteIDParam := c.Param("id")
	noteID, err := strconv.Atoi(noteIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid note ID")
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

	var note models.Note
	// Ambil catatan dari database
	if err := models.DB.Where("id = ? AND project_id = ?", noteID, project.ID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Hapus catatan dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to delete note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	utils.Logger.Infof("Note deleted successfully: NoteID %d in ProjectID %d", note.ID, project.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Note deleted successfully"})
}
