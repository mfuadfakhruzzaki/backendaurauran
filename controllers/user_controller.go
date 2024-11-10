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
	// Get user_id from context set by AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var user models.User
	// Retrieve user data from the database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user profile: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve profile")
		return
	}

	// Prepare response data without sending password
	responseData := gin.H{
		"id":                user.ID,
		"username":          user.Username,
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
	Username string `json:"username" binding:"omitempty"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// UpdateProfile handles updating the user's profile
func UpdateProfile(c *gin.Context) {
	// Get user_id from context set by AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var req UpdateProfileRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// If no fields to update
	if req.Username == "" && req.Email == "" && req.Password == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "No fields to update")
		return
	}

	var user models.User
	// Retrieve user data from the database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user for update: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// Update username if provided
	if req.Username != "" && req.Username != user.Username {
		user.Username = req.Username
	}

	// Update email if provided
	if req.Email != "" && req.Email != user.Email {
		user.Email = req.Email
		user.IsEmailVerified = false // Mark email as unverified
	}

	// Update password if provided
	if req.Password != "" {
		user.Password = req.Password // Password will be hashed by GORM hooks
	}

	// Save changes to the database
	if err := models.DB.Save(&user).Error; err != nil {
		if isUniqueConstraintError(err) {
			utils.ErrorResponse(c, http.StatusConflict, "Email or username already exists")
			return
		}
		utils.Logger.Errorf("Failed to update user profile: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	// If email was changed, send verification email
	if req.Email != "" && req.Email != user.Email {
		// Create new email verification token
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
			ExpiresAt: time.Now().Add(24 * time.Hour), // Token valid for 24 hours
		}

		if err := models.DB.Create(&emailToken).Error; err != nil {
			utils.Logger.Errorf("Failed to save verification token: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
			return
		}

		// Send new verification email
		emailService := utils.NewEmailService()
		if err := emailService.SendVerificationEmail(user.Email, verifyToken); err != nil {
			utils.Logger.Errorf("Failed to send verification email: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send verification email")
			return
		}

		utils.Logger.Infof("Verification email sent to: %s", user.Email)
	}

	utils.Logger.Infof("User profile updated successfully: UserID %d", user.ID)

	// Prepare response data without sending password
	responseData := gin.H{
		"id":                user.ID,
		"username":          user.Username,
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
	// Get user_id from context set by AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		utils.Logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}

	var user models.User
	// Retrieve user data from the database
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve user for deletion: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	// Delete user from the database (soft delete if using gorm.DeletedAt)
	if err := models.DB.Delete(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to delete user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	utils.Logger.Infof("User account deleted successfully: UserID %d", user.ID)

	// Send success response
	utils.SuccessResponse(c, gin.H{"message": "Account deleted successfully"})
}

// isUniqueConstraintError checks if an error is due to a unique constraint violation in PostgreSQL
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
