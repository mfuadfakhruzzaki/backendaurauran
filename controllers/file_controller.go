// controllers/file_controller.go
package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/storage"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// FileController handles file-related operations
type FileController struct {
	DB             *gorm.DB
	StorageService storage.StorageService
	BucketName     string
}

// NewFileController creates a new FileController instance
func NewFileController(db *gorm.DB, storageService storage.StorageService, bucketName string) *FileController {
	return &FileController{
		DB:             db,
		StorageService: storageService,
		BucketName:     bucketName,
	}
}

// UploadFile handles uploading a file to S3 and saving metadata to the database
func (fc *FileController) UploadFile(c *gin.Context) {
	// Get project_id from URL parameters and convert to int
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil || projectID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the project exists
	var project models.Project
	if err := fc.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
		return
	}

	// Validate file type if necessary
	if !fc.isValidFileType(file.Header.Get("Content-Type")) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file type")
		return
	}

	// Validate file size (e.g., max 5MB)
	const MaxFileSize = 5 << 20 // 5MB
	if file.Size > MaxFileSize {
		utils.ErrorResponse(c, http.StatusBadRequest, "File size exceeds the limit of 5MB")
		return
	}

	// Open the file
	f, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer f.Close()

	// Generate unique object name
	objectName := storage.GenerateUniqueObjectName(file.Filename)

	// Upload to S3
	fileURL, err := fc.StorageService.UploadFile(context.Background(), fc.BucketName, objectName, f, file.Header.Get("Content-Type"))
	if err != nil {
		utils.Logger.Errorf("Failed to upload file to storage: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	// Get user from context set by AuthMiddleware
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

	// Save file metadata to database
	fileModel := models.File{
		ProjectID:  uint(projectID),
		Filename:   file.Filename,
		FileURL:    fileURL,
		FileType:   file.Header.Get("Content-Type"),
		FileSize:   file.Size,
		UploadedBy: user.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Logging before saving
	utils.Logger.Infof("Saving file metadata: %+v", fileModel)

	if err := fc.DB.Create(&fileModel).Error; err != nil {
		utils.Logger.Errorf("Failed to save file metadata to database: %v", err)
		// If saving metadata fails, delete the file from storage
		if delErr := fc.StorageService.DeleteFile(context.Background(), fc.BucketName, objectName); delErr != nil {
			utils.Logger.Errorf("Failed to delete file from storage after DB failure: %v", delErr)
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file metadata")
		return
	}

	utils.Logger.Infof("File uploaded successfully: FileID %d in ProjectID %d by UserID %d", fileModel.ID, projectID, user.ID)

	// Prepare response data
	responseData := gin.H{
		"id":          fileModel.ID,
		"project_id":  fileModel.ProjectID,
		"filename":    fileModel.Filename,
		"file_url":    fileModel.FileURL,
		"file_type":   fileModel.FileType,
		"file_size":   fileModel.FileSize,
		"uploaded_by": fileModel.UploadedBy,
		"created_at":  fileModel.CreatedAt,
		"updated_at":  fileModel.UpdatedAt,
	}

	// Send success response
	utils.CreatedResponse(c, responseData)
}

// ListFiles handles retrieving all files within a project
func (fc *FileController) ListFiles(c *gin.Context) {
	// Get project_id from URL parameters and convert to int
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil || projectID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Check if the project exists
	var project models.Project
	if err := fc.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Get files from database
	var files []models.File
	if err := fc.DB.Where("project_id = ?", projectID).Order("created_at desc").Find(&files).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve files: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	// Prepare response data
	var responseData []gin.H
	for _, file := range files {
		responseData = append(responseData, gin.H{
			"id":          file.ID,
			"project_id":  file.ProjectID,
			"filename":    file.Filename,
			"file_url":    file.FileURL,
			"file_type":   file.FileType,
			"file_size":   file.FileSize,
			"uploaded_by": file.UploadedBy,
			"created_at":  file.CreatedAt,
			"updated_at":  file.UpdatedAt,
		})
	}

	// Send success response
	utils.SuccessResponse(c, responseData)
}

// DownloadFile handles downloading a file from S3
func (fc *FileController) DownloadFile(c *gin.Context) {
	// Get project_id and file_id from URL parameters and convert to int
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil || projectID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	fileIDParam := c.Param("id")
	fileID, err := strconv.Atoi(fileIDParam)
	if err != nil || fileID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	// Check if the project exists
	var project models.Project
	if err := fc.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Get file from database
	var file models.File
	if err := fc.DB.Where("id = ? AND project_id = ?", fileID, projectID).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "File not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve file: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve file")
		return
	}

	// Generate presigned URL
	presignedURL, err := fc.generatePresignedURL(context.Background(), fc.BucketName, getObjectNameFromURL(file.FileURL))
	if err != nil {
		utils.Logger.Errorf("Failed to generate presigned URL: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate presigned URL")
		return
	}

	// Redirect user to the presigned URL
	c.Redirect(http.StatusFound, presignedURL)
}

// DeleteFile handles deleting a file from S3 and the database
func (fc *FileController) DeleteFile(c *gin.Context) {
	// Get project_id and file_id from URL parameters and convert to int
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil || projectID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	fileIDParam := c.Param("id")
	fileID, err := strconv.Atoi(fileIDParam)
	if err != nil || fileID <= 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	// Check if the project exists
	var project models.Project
	if err := fc.DB.First(&project, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Get file from database
	var file models.File
	if err := fc.DB.Where("id = ? AND project_id = ?", fileID, projectID).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "File not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve file: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve file")
		return
	}

	// Get user from context set by AuthMiddleware
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

	// Authorization: Ensure that the user deleting the file is the project owner or the uploader
	if file.UploadedBy != user.ID && project.OwnerID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this file")
		return
	}

	// Extract object name from file URL
	objectName := getObjectNameFromURL(file.FileURL)
	if objectName == "" {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid file URL")
		return
	}

	// Delete the file from storage
	if err := fc.StorageService.DeleteFile(context.Background(), fc.BucketName, objectName); err != nil {
		utils.Logger.Errorf("Failed to delete file from storage: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete file from storage")
		return
	}

	// Delete file metadata from database
	if err := fc.DB.Delete(&file).Error; err != nil {
		utils.Logger.Errorf("Failed to delete file metadata from database: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete file metadata")
		return
	}

	utils.Logger.Infof("File deleted successfully: FileID %d for ProjectID %d by UserID %d", file.ID, projectID, user.ID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "File deleted successfully"})
}

// isValidFileType validates the file type
func (fc *FileController) isValidFileType(fileType string) bool {
	// Add valid file types here
	validFileTypes := map[string]bool{
		"image/jpeg":       true,
		"image/png":        true,
		"application/pdf":  true,
		"video/mp4":        true,
		// Tambahkan tipe file yang diizinkan lainnya
	}
	return validFileTypes[fileType]
}

// generatePresignedURL generates a presigned URL for accessing the file
func (fc *FileController) generatePresignedURL(ctx context.Context, bucketName, objectName string) (string, error) {
	presignedURL, err := fc.StorageService.GeneratePresignedURL(ctx, bucketName, objectName, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}
	return presignedURL, nil
}

// getObjectNameFromURL extracts the object name from the file URL
func getObjectNameFromURL(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return ""
	}
	// Extract the object part from the URL
	// For example, https://bucket-name.s3.region.amazonaws.com/object-name
	segments := strings.Split(parsedURL.Path, "/")
	if len(segments) < 2 {
		return ""
	}
	// Rejoin the object parts in case the object name contains '/'
	objectName := strings.Join(segments[1:], "/")
	return objectName
}
