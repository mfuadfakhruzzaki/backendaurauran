// controllers/task_controller.go
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

// CreateTaskRequest merepresentasikan struktur request untuk membuat tugas baru
type CreateTaskRequest struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description" binding:"required"`
	Priority    models.TaskPriority `json:"priority" binding:"required,oneof=low medium high"`
	Status      models.TaskStatus   `json:"status" binding:"required,oneof=pending in_progress completed cancelled"`
	Deadline    *time.Time         `json:"deadline" binding:"omitempty"`
	AssignedTo  *uint              `json:"assigned_to" binding:"omitempty"`
}

// UpdateTaskRequest merepresentasikan struktur request untuk memperbarui tugas
type UpdateTaskRequest struct {
	Title       *string             `json:"title" binding:"omitempty"`
	Description *string             `json:"description" binding:"omitempty"`
	Priority    *models.TaskPriority `json:"priority" binding:"omitempty,oneof=low medium high"`
	Status      *models.TaskStatus   `json:"status" binding:"omitempty,oneof=pending in_progress completed cancelled"`
	Deadline    *time.Time         `json:"deadline" binding:"omitempty"`
	AssignedTo  *uint              `json:"assigned_to" binding:"omitempty"`
}

// CreateTask menangani pembuatan tugas baru dalam proyek tertentu
func CreateTask(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var req CreateTaskRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Jika AssignedTo disediakan, cek apakah pengguna yang ditugaskan ada
	if req.AssignedTo != nil {
		var user models.User
		if err := models.DB.First(&user, *req.AssignedTo).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusNotFound, "Assigned user not found")
				return
			}
			utils.Logger.Errorf("Failed to retrieve assigned user: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve assigned user")
			return
		}
	}

	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Buat instance Task baru
	task := models.Task{
		ProjectID:   uint(projectID),
		AssignedTo:  req.AssignedTo, // *uint, sekarang sesuai dengan model
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      req.Status,
		Deadline:    req.Deadline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Simpan task ke database
	if err := models.DB.Create(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to create task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create task")
		return
	}

	utils.Logger.Infof("Task created successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, currentUserID.(uint))

	// Kirim respons sukses dengan data task
	utils.CreatedResponse(c, gin.H{
		"id":          task.ID,
		"project_id":  task.ProjectID,
		"assigned_to": task.AssignedTo,
		"title":       task.Title,
		"description": task.Description,
		"priority":    task.Priority,
		"status":      task.Status,
		"deadline":    task.Deadline,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
	})
}

// ListTasks menangani pengambilan semua tugas dalam proyek tertentu
func ListTasks(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var tasks []models.Task
	// Ambil semua tugas untuk proyek, urutkan dari yang terbaru
	if err := models.DB.Where("project_id = ?", projectID).Order("created_at desc").Find(&tasks).Error; err != nil {
		utils.Logger.Errorf("Failed to retrieve tasks: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tasks")
		return
	}

	// Siapkan data respons
	var responseData []gin.H
	for _, task := range tasks {
		responseData = append(responseData, gin.H{
			"id":          task.ID,
			"project_id":  task.ProjectID,
			"assigned_to": task.AssignedTo,
			"title":       task.Title,
			"description": task.Description,
			"priority":    task.Priority,
			"status":      task.Status,
			"deadline":    task.Deadline,
			"created_at":  task.CreatedAt,
			"updated_at":  task.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, responseData)
}

// GetTask menangani pengambilan detail tugas tertentu dalam proyek
func GetTask(c *gin.Context) {
	// Ambil project_id dan task_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("id")
	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var task models.Task
	// Ambil tugas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", taskID, projectID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// Siapkan data respons
	responseData := gin.H{
		"id":          task.ID,
		"project_id":  task.ProjectID,
		"assigned_to": task.AssignedTo,
		"title":       task.Title,
		"description": task.Description,
		"priority":    task.Priority,
		"status":      task.Status,
		"deadline":    task.Deadline,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateTask menangani pembaruan tugas tertentu dalam proyek
func UpdateTask(c *gin.Context) {
	// Ambil project_id dan task_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("id")
	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	var req UpdateTaskRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek apakah ada setidaknya satu field yang diubah
	if req.Title == nil && req.Description == nil && req.Priority == nil && req.Status == nil && req.Deadline == nil && req.AssignedTo == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var task models.Task
	// Ambil tugas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", taskID, projectID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Otorisasi: Pastikan bahwa pengguna yang mengupdate tugas adalah pemilik proyek atau kolaborator dengan hak akses yang sesuai
	// Implementasi otorisasi ini tergantung pada logika bisnis Anda
	// Misalnya, Anda dapat memeriksa apakah currentUserID adalah pemilik proyek atau memiliki peran tertentu

	// Jika AssignedTo disediakan, cek apakah pengguna yang ditugaskan ada
	if req.AssignedTo != nil {
		var user models.User
		if err := models.DB.First(&user, *req.AssignedTo).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.ErrorResponse(c, http.StatusNotFound, "Assigned user not found")
				return
			}
			utils.Logger.Errorf("Failed to retrieve assigned user: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve assigned user")
			return
		}
		task.AssignedTo = req.AssignedTo
	}

	// Update field jika disediakan
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}
	task.UpdatedAt = time.Now()

	// Simpan perubahan ke database
	if err := models.DB.Save(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to update task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update task")
		return
	}

	utils.Logger.Infof("Task updated successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, currentUserID.(uint))

	// Siapkan data respons
	responseData := gin.H{
		"id":          task.ID,
		"project_id":  task.ProjectID,
		"assigned_to": task.AssignedTo,
		"title":       task.Title,
		"description": task.Description,
		"priority":    task.Priority,
		"status":      task.Status,
		"deadline":    task.Deadline,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteTask menangani penghapusan tugas tertentu dalam proyek
func DeleteTask(c *gin.Context) {
	// Ambil project_id dan task_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	taskIDParam := c.Param("id")
	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Cek apakah proyek yang dituju ada
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

	var task models.Task
	// Ambil tugas dari database
	if err := models.DB.Where("id = ? AND project_id = ?", taskID, projectID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Task not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve task")
		return
	}

	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	// Otorisasi: Pastikan bahwa pengguna yang menghapus tugas adalah pemilik proyek atau kolaborator dengan hak akses yang sesuai
	// Implementasi otorisasi ini tergantung pada logika bisnis Anda
	// Misalnya, Anda dapat memeriksa apakah currentUserID adalah pemilik proyek atau memiliki peran tertentu

	// Hapus tugas dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&task).Error; err != nil {
		utils.Logger.Errorf("Failed to delete task: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete task")
		return
	}

	utils.Logger.Infof("Task deleted successfully: TaskID %d for ProjectID %d by UserID %d", task.ID, projectID, currentUserID.(uint))

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Task deleted successfully"})
}
