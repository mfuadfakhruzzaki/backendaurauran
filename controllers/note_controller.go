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

	// Check if the project exists and the user has access
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create note")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var req CreateNoteRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create a new Note instance
	note := models.Note{
		ProjectID: uint(projectID),
		UserID:    user.ID,
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save note to database
	if err := models.DB.Create(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to create note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create note")
		return
	}

	utils.Logger.Infof("Note created successfully: NoteID %d in ProjectID %d by UserID %d", note.ID, projectID, user.ID)

	// Prepare response data
	responseData := gin.H{
		"id":          note.ID,
		"project_id":  note.ProjectID,
		"user_id":     note.UserID,
		"content":     note.Content,
		"created_at":  note.CreatedAt,
		"updated_at":  note.UpdatedAt,
	}

	// Send success response with note data
	utils.CreatedResponse(c, responseData)
}

// ListNotes handles retrieving all notes within a project
func ListNotes(c *gin.Context) {
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

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var notes []models.Note
	// Retrieve all notes for the project, including the User who created each note
	if err := models.DB.Where("project_id = ?", uint(projectID)).
		Preload("User").
		Order("created_at desc").
		Find(&notes).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve notes: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}

	// Prepare response data
	var responseData []gin.H
	for _, note := range notes {
		responseData = append(responseData, gin.H{
			"id":          note.ID,
			"project_id":  note.ProjectID,
			"user_id":     note.UserID,
			"content":     note.Content,
			"created_at":  note.CreatedAt,
			"updated_at":  note.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetNote handles retrieving a single note by ID within a project
func GetNote(c *gin.Context) {
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

	// Retrieve project_id and note_id from URL parameters
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

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var note models.Note
	// Retrieve the note from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(noteID), uint(projectID)).
		Preload("User").
		First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Prepare response data
	responseData := gin.H{
		"id":          note.ID,
		"project_id":  note.ProjectID,
		"user_id":     note.UserID,
		"content":     note.Content,
		"created_at":  note.CreatedAt,
		"updated_at":  note.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateNote handles updating a note within a project
func UpdateNote(c *gin.Context) {
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

	// Retrieve project_id and note_id from URL parameters
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
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if at least one field is provided for update
	if req.Content == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update note")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var note models.Note
	// Retrieve the note from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(noteID), uint(projectID)).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Update fields if provided
	if req.Content != "" {
		note.Content = req.Content
	}
	note.UpdatedAt = time.Now()

	// Save changes to the database
	if err := models.DB.Save(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to update note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update note")
		return
	}

	utils.Logger.Infof("Note updated successfully: NoteID %d in ProjectID %d by UserID %d", note.ID, projectID, user.ID)

	// Prepare response data
	responseData := gin.H{
		"id":          note.ID,
		"project_id":  note.ProjectID,
		"user_id":     note.UserID,
		"content":     note.Content,
		"created_at":  note.CreatedAt,
		"updated_at":  note.UpdatedAt,
	}

	// Send success response
	utils.SuccessResponse(c, responseData)
}

// DeleteNote handles deleting a note within a project
func DeleteNote(c *gin.Context) {
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

	// Retrieve project_id and note_id from URL parameters
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

	// Check if the user has access to the project
	hasAccess, err := models.UserHasAccessToProject(user.ID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	var note models.Note
	// Retrieve the note from the database
	if err := models.DB.Where("id = ? AND project_id = ?", uint(noteID), uint(projectID)).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	// Delete the note from the database (soft delete if using gorm.DeletedAt)
	if err := models.DB.Delete(&note).Error; err != nil {
		utils.Logger.Errorf("Failed to delete note: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	utils.Logger.Infof("Note deleted successfully: NoteID %d in ProjectID %d by UserID %d", note.ID, projectID, user.ID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "Note deleted successfully"})
}
