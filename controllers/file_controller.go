// controllers/file_controller.go
package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/storage"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// FileController menangani operasi terkait file
type FileController struct {
	DB             *gorm.DB
	StorageService storage.StorageService
	BucketName     string
}

// NewFileController membuat instance baru FileController
func NewFileController(db *gorm.DB, storageService storage.StorageService, bucketName string) *FileController {
	return &FileController{
		DB:             db,
		StorageService: storageService,
		BucketName:     bucketName,
	}
}

// UploadFileRequest merepresentasikan struktur request untuk mengupload file
type UploadFileRequest struct {
	// Anda bisa menambahkan field tambahan jika diperlukan
}

// UploadFile menangani upload file ke Google Cloud Storage dan menyimpan metadata ke database
func (fc *FileController) UploadFile(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	// Form File
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
		return
	}

	// Validasi tipe file jika diperlukan
	// e.g., hanya menerima gambar
	if !isValidFileType(file.Header.Get("Content-Type")) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file type")
		return
	}

	// Validasi ukuran file (misalnya maksimal 5MB)
	const MaxFileSize = 5 << 20 // 5MB
	if file.Size > MaxFileSize {
		utils.ErrorResponse(c, http.StatusBadRequest, "File size exceeds the limit of 5MB")
		return
	}

	// Buka file
	f, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer f.Close()

	// Generate nama objek unik
	objectName := storage.GenerateUniqueObjectName(file.Filename)

	// Upload ke GCS
	fileURL, err := fc.StorageService.UploadFile(context.Background(), fc.BucketName, objectName, f, file.Header.Get("Content-Type"))
	if err != nil {
		utils.Logger.Errorf("Failed to upload file to storage: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file")
		return
	}

	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Simpan metadata file ke database
	fileModel := models.File{
		ProjectID:  uint(projectID),
		FileName:   file.Filename,
		FileURL:    fileURL,
		FileType:   file.Header.Get("Content-Type"),
		FileSize:   file.Size,
		UploadedBy: currentUserID.(uint),
	}

	if err := fc.DB.Create(&fileModel).Error; err != nil {
		utils.Logger.Errorf("Failed to save file metadata to database: %v", err)
		// Jika penyimpanan metadata gagal, hapus file dari storage
		fc.StorageService.DeleteFile(context.Background(), fc.BucketName, objectName)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file metadata")
		return
	}

	// Kirim respons sukses
	utils.CreatedResponse(c, gin.H{
		"id":          fileModel.ID,
		"project_id":  fileModel.ProjectID,
		"file_name":   fileModel.FileName,
		"file_url":    fileModel.FileURL,
		"file_type":   fileModel.FileType,
		"file_size":   fileModel.FileSize,
		"uploaded_by": fileModel.UploadedBy,
		"created_at":  fileModel.CreatedAt,
		"updated_at":  fileModel.UpdatedAt,
	})
}

// isValidFileType memeriksa apakah tipe file diizinkan
func isValidFileType(contentType string) bool {
	// Contoh: hanya menerima gambar
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	return allowedTypes[contentType]
}

// ListFiles menangani pengambilan semua file dalam proyek tertentu
func (fc *FileController) ListFiles(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var files []models.File
	if err := fc.DB.Where("project_id = ?", projectID).Order("created_at desc").Find(&files).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve files: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve files")
		return
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, file := range files {
		responseData = append(responseData, gin.H{
			"id":          file.ID,
			"project_id":  file.ProjectID,
			"file_name":   file.FileName,
			"file_url":    file.FileURL,
			"file_type":   file.FileType,
			"file_size":   file.FileSize,
			"uploaded_by": file.UploadedBy,
			"created_at":  file.CreatedAt,
			"updated_at":  file.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// DownloadFile menangani pengunduhan file dari GCS
func (fc *FileController) DownloadFile(c *gin.Context) {
	// Ambil project_id dan file_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	fileIDParam := c.Param("id")
	fileID, err := strconv.Atoi(fileIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	// Ambil file dari database
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

	// Redirect pengguna ke presigned URL
	c.Redirect(http.StatusFound, presignedURL)
}

// generatePresignedURL membuat presigned URL untuk mengakses file
func (fc *FileController) generatePresignedURL(ctx context.Context, bucketName, objectName string) (string, error) {
	// Menggunakan GCSStorageService untuk membuat signed URL
	gcsService, ok := fc.StorageService.(*storage.GCSStorageService)
	if !ok {
		return "", fmt.Errorf("invalid storage service type")
	}

	// Atur durasi signed URL (misalnya 15 menit)
	expiration := 15 * time.Minute

	// Generate signed URL
	signedURL, err := gcsService.GeneratePresignedURL(ctx, bucketName, objectName, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %v", err)
	}

	return signedURL, nil
}

// getObjectNameFromURL mengambil nama objek dari URL file
func getObjectNameFromURL(fileURL string) string {
	// URL format: https://storage.googleapis.com/<bucket>/<object>
	var bucketName, objectName string
	// Gunakan fmt.Sscanf dengan format yang benar
	_, err := fmt.Sscanf(fileURL, "https://storage.googleapis.com/%s/%s", &bucketName, &objectName)
	if err != nil {
		return ""
	}
	return objectName
}

// DeleteFile menangani penghapusan file dari GCS dan database
func (fc *FileController) DeleteFile(c *gin.Context) {
	// Ambil project_id dan file_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	fileIDParam := c.Param("id")
	fileID, err := strconv.Atoi(fileIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	// Ambil file dari database
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

	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Otorisasi: Pastikan bahwa pengguna yang menghapus file adalah pemilik proyek atau yang mengupload file
	if file.UploadedBy != currentUserID.(uint) && project.OwnerID != currentUserID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this file")
		return
	}

	// Delete file dari storage
	objectName := getObjectNameFromURL(file.FileURL)
	if objectName == "" {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid file URL")
		return
	}

	if err := fc.StorageService.DeleteFile(context.Background(), fc.BucketName, objectName); err != nil {
		utils.Logger.Errorf("Failed to delete file from storage: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete file from storage")
		return
	}

	// Hapus metadata file dari database
	if err := fc.DB.Delete(&file).Error; err != nil {
		utils.Logger.Errorf("Failed to delete file metadata from database: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete file metadata")
		return
	}

	utils.Logger.Infof("File deleted successfully: FileID %d for ProjectID %d by UserID %d", file.ID, projectID, currentUserID.(uint))

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "File deleted successfully"})
}
