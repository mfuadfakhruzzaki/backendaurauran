// controllers/user_controller.go
package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// GetProfile handles retrieving the user's profile
func GetProfile(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var user models.User
	// Ambil data user dari database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user profile: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}

	// Siapkan data respons tanpa mengirimkan password
	responseData := gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"role":              user.Role,
		"is_email_verified": user.IsEmailVerified,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// UpdateProfileRequest represents the request structure for updating user profile
type UpdateProfileRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// UpdateProfile handles updating the user's profile
func UpdateProfile(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var req UpdateProfileRequest
	// Bind JSON request ke struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Jika tidak ada field yang diubah
	if req.Email == "" && req.Password == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	var user models.User
	// Ambil data user dari database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user for update: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Update email jika disediakan
	if req.Email != "" && req.Email != user.Email {
		user.Email = req.Email
		user.IsEmailVerified = false // Tandai ulang verifikasi email
	}

	// Update password jika disediakan
	if req.Password != "" {
		user.Password = req.Password // Password akan di-hash oleh GORM hooks
	}

	// Simpan perubahan ke database
	if err := models.DB.Save(&user).Error; err != nil {
		if isUniqueConstraintError(err) {
			utils.ErrorResponse(c, http.StatusConflict, "Email already exists")
			return
		}
		utils.Logger.Errorf("Failed to update user profile: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Jika email diubah, kirim email verifikasi ulang
	if req.Email != "" && req.Email != user.Email {
		// Buat token verifikasi email baru
		verifyToken, err := utils.GenerateRandomToken(32)
		if err != nil {
			utils.Logger.Errorf("Failed to generate verification token: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
			return
		}

		emailToken := models.Token{
			UserID:    user.ID,
			Token:     verifyToken,
			Type:      models.TokenTypeEmailVerify,
			ExpiresAt: time.Now().Add(24 * time.Hour), // Token valid selama 24 jam
		}

		if err := models.DB.Create(&emailToken).Error; err != nil {
			utils.Logger.Errorf("Failed to save verification token: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
			return
		}

		// Kirim email verifikasi baru
		emailService := utils.NewEmailService()
		if err := emailService.SendVerificationEmail(user.Email, verifyToken); err != nil {
			utils.Logger.Errorf("Failed to send verification email: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send verification email")
			return
		}

		utils.Logger.Infof("Verification email sent to: %s", user.Email)
	}

	utils.Logger.Infof("User profile updated successfully: UserID %d", user.ID)

	// Siapkan data respons tanpa mengirimkan password
	responseData := gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"role":              user.Role,
		"is_email_verified": user.IsEmailVerified,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	}

	utils.SuccessResponse(c, responseData)
}

// DeleteProfile handles deleting the user's account (Optional)
func DeleteProfile(c *gin.Context) {
	// Ambil user_id dari context yang sudah di-set oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var user models.User
	// Ambil data user dari database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user for deletion: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	// Hapus user dari database (soft delete jika menggunakan gorm.DeletedAt)
	if err := models.DB.Delete(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to delete user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	utils.Logger.Infof("User account deleted successfully: UserID %d", user.ID)

	// Kirim respons sukses
	utils.SuccessResponse(c, gin.H{"message": "Account deleted successfully"})
}

// isUniqueConstraintError memeriksa apakah error berasal dari pelanggaran unique constraint PostgreSQL
func isUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }
    if pqErr, ok := err.(*pq.Error); ok {
        return pqErr.Code == "23505"
    }
    return false
}
