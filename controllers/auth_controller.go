// controllers/auth_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
)

// RegisterRequest represents the request structure for user registration
type RegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	InvitationCode string `json:"invitation_code"`
}

// LoginRequest represents the request structure for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RequestPasswordResetRequest represents the request structure for requesting a password reset
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the request structure for resetting the password via API
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPasswordFormRequest represents the form data for resetting the password via HTML form
type ResetPasswordFormRequest struct {
	Token           string `form:"token" binding:"required"`
	NewPassword     string `form:"new_password" binding:"required,min=6"`
	ConfirmPassword string `form:"confirm_password" binding:"required,min=6"`
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Determine user role based on invitation code
	role := models.RoleMember // Default role
	if req.InvitationCode == "GWEHADMIN" {
		role = models.RoleAdmin
	} else if req.InvitationCode == "GWEHMANAGER" {
		role = models.RoleManager
	}

	// Check if user with the same email or username already exists
	var existingUser models.User
	if err := models.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		if existingUser.Email == req.Email {
			utils.ErrorResponse(c, http.StatusConflict, "Email already registered")
		} else if existingUser.Username == req.Username {
			utils.ErrorResponse(c, http.StatusConflict, "Username already taken")
		}
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.Logger.Errorf("Failed to check existing user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Create new user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Password will be hashed by GORM hook
		Role:     role,
	}

	if err := models.DB.Create(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to create user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Generate email verification token
	verifyToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		utils.Logger.Errorf("Failed to generate verification token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Send verification email
	emailService := utils.NewEmailService()
	verifyURL := fmt.Sprint(verifyToken)
	if err := emailService.SendVerificationEmail(user.Email, verifyURL); err != nil {
		utils.Logger.Errorf("Failed to send verification email: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send verification email")
		return
	}

	utils.Logger.Infof("User registered successfully: %s with role %s", user.Email, user.Role)

	// Send success response
	utils.CreatedResponse(c, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	// Find user by email
	if err := models.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		utils.Logger.Errorf("Failed to find user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	// Check if password matches
	if !user.ComparePassword(req.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Check if email is verified
	if !user.IsEmailVerified {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email not verified")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		utils.Logger.Errorf("Failed to generate JWT: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	utils.Logger.Infof("User logged in successfully: %s", user.Email)

	// Send success response with token
	utils.SuccessResponse(c, gin.H{
		"token": token,
	})
}

// Logout handles user logout by blacklisting the JWT token
func Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Authorization header required")
		return
	}

	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid authorization header format")
		return
	}
	tokenStr := parts[1]

	// Parse JWT token to get claims
	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		utils.Logger.Errorf("Failed to parse JWT token: %v", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Add token to blacklist with UserID from claims
	blacklistToken := models.Token{
		UserID:    claims.UserID,
		Token:     tokenStr,
		Type:      models.TokenTypeJWTBlacklist,
		ExpiresAt: time.Now().Add(time.Hour * 24), // Adjust blacklist duration
	}

	if err := models.DB.Create(&blacklistToken).Error; err != nil {
		utils.Logger.Errorf("Failed to blacklist token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to logout")
		return
	}

	utils.Logger.Infof("Token blacklisted successfully for user ID: %d", claims.UserID)

	// Send success response
	utils.SuccessResponse(c, gin.H{
		"message": "Successfully logged out",
	})
}

// VerifyEmail handles email verification
func VerifyEmail(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Render failure page with message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", tokenStr, models.TokenTypeEmailVerify).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render failure page with message
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		utils.Logger.Errorf("Failed to verify email token: %v", err)
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render failure page with expired token message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var user models.User
	if err := models.DB.First(&user, token.UserID).Error; err != nil {
		utils.Logger.Errorf("User not found for email verification: %v", err)
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	user.IsEmailVerified = true

	if err := models.DB.Save(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to update user email verification: %v", err)
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Delete the token after verification
	if err := models.DB.Delete(&token).Error; err != nil {
		utils.Logger.Errorf("Failed to delete email verification token: %v", err)
		// Do not return error since email is already verified
	}

	utils.Logger.Infof("Email verified successfully for user: %s", user.Email)

	// Render email verification success page
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(successVerifyHTML))
}

// RequestPasswordReset handles password reset requests
func RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	// Find user by email
	if err := models.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Do not reveal whether the email exists or not for security
			utils.SuccessResponse(c, gin.H{
				"message": "If the email is registered, a password reset link has been sent.",
			})
			return
		}
		utils.Logger.Errorf("Failed to find user for password reset: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to request password reset")
		return
	}

	// Generate password reset token
	resetToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		utils.Logger.Errorf("Failed to generate password reset token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to request password reset")
		return
	}

	passwordResetToken := models.Token{
		UserID:    user.ID,
		Token:     resetToken,
		Type:      models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token valid for 24 hours
	}

	if err := models.DB.Create(&passwordResetToken).Error; err != nil {
		utils.Logger.Errorf("Failed to save password reset token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to request password reset")
		return
	}

	// Send password reset email
	emailService := utils.NewEmailService()
	resetURL := fmt.Sprint(resetToken)
	if err := emailService.SendResetPasswordEmail(user.Email, resetURL); err != nil {
		utils.Logger.Errorf("Failed to send password reset email: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to send password reset email")
		return
	}

	utils.Logger.Infof("Password reset email sent to: %s", user.Email)

	// Send success response
	utils.SuccessResponse(c, gin.H{
		"message": "If the email is registered, a password reset link has been sent.",
	})
}

// ResetPasswordForm handles rendering the reset password form (GET request)
func ResetPasswordForm(c *gin.Context) {
	// Get token from query parameter
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Render failure page with message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", tokenStr, models.TokenTypePasswordReset).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render failure page with message
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		utils.Logger.Errorf("Failed to verify reset token: %v", err)
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render failure page with expired token message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Render reset password page with form
	// You can embed the token in the form as a hidden field
	htmlContent := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Reset Password</title>
		<style>
			body {
				background-color: #f2f2f2;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				font-family: Arial, sans-serif;
				margin: 0;
			}
			.container {
				background-color: #ffffff;
				padding: 40px;
				box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
				border-radius: 8px;
				text-align: center;
			}
			.container h1 {
				color: #007bff;
				font-size: 2em;
				margin-bottom: 20px;
			}
			.container form {
				display: flex;
				flex-direction: column;
			}
			.container input {
				padding: 10px;
				margin: 10px 0;
				border: 1px solid #ccc;
				border-radius: 4px;
				font-size: 1em;
			}
			.container button {
				padding: 10px;
				background-color: #28a745;
				color: #ffffff;
				border: none;
				border-radius: 4px;
				font-size: 1em;
				cursor: pointer;
				transition: background-color 0.3s ease;
			}
			.container button:hover {
				background-color: #218838;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Reset Your Password</h1>
			<form action="/auth/reset-password" method="POST">
				<input type="hidden" name="token" value="{{.Token}}">
				<input type="password" name="new_password" placeholder="Enter New Password" required minlength="6">
				<input type="password" name="confirm_password" placeholder="Confirm New Password" required minlength="6">
				<button type="submit">Reset Password</button>
			</form>
		</div>
	</body>
	</html>
	`

	// Replace {{.Token}} with the actual token
	finalHTML := strings.Replace(htmlContent, "{{.Token}}", tokenStr, 1)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(finalHTML))
}

// ResetPassword handles resetting the user's password using the provided token (POST request)
func ResetPassword(c *gin.Context) {
	var req ResetPasswordFormRequest
	// Bind form data to struct
	if err := c.ShouldBind(&req); err != nil {
		// Render failure page with message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>"+"Invalid input: "+err.Error()+"</p>"))
		return
	}

	// Validate that new_password and confirm_password match
	if req.NewPassword != req.ConfirmPassword {
		// Render failure page with message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>Passwords do not match</p>"))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", req.Token, models.TokenTypePasswordReset).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render failure page with message
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render failure page with expired token message
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var user models.User
	if err := models.DB.First(&user, token.UserID).Error; err != nil {
		// Render generic failure page
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Manually hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		// Render failure page with message
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Update password with the new hashed password
	user.Password = string(hashedPassword)

	if err := models.DB.Save(&user).Error; err != nil {
		// Render failure page with message
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Delete the token after use
	if err := models.DB.Delete(&token).Error; err != nil {
		// Do not return error since password is already changed
	}

	// Render password reset success page
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(successResetHTML))
}

// ResetPasswordAPI handles resetting the user's password via API (optional)
func ResetPasswordAPI(c *gin.Context) {
	var req ResetPasswordRequest
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", req.Token, models.TokenTypePasswordReset).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid or expired token")
			return
		}
		utils.Logger.Errorf("Failed to verify reset token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Token has expired")
		return
	}

	var user models.User
	if err := models.DB.First(&user, token.UserID).Error; err != nil {
		utils.Logger.Errorf("User not found for reset token: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "User not found")
		return
	}

	// Update password (pastikan password di-hash oleh GORM hooks atau lakukan hashing di sini)
	user.Password = req.NewPassword

	if err := models.DB.Save(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to update user password: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	// Delete the token after use
	if err := models.DB.Delete(&token).Error; err != nil {
		utils.Logger.Errorf("Failed to delete reset token: %v", err)
		// Tidak mengembalikan error karena password sudah diubah
	}

	utils.Logger.Infof("Password reset successfully for user: %s", user.Email)

	// Send success response
	utils.SuccessResponse(c, gin.H{
		"message": "Password reset successfully",
	})
}

// successVerifyHTML defines the HTML content for successful email verification
const successVerifyHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Email Verification Successful</title>
	<style>
		body {
			background-color: #d4edda;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
			font-family: Arial, sans-serif;
			margin: 0;
		}
		.container {
			background-color: #ffffff;
			padding: 40px;
			box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
			border-radius: 8px;
			text-align: center;
		}
		.container h1 {
			color: #155724;
			font-size: 2em;
			margin-bottom: 20px;
		}
		.container p {
			font-size: 1em;
			color: #333333;
			margin-bottom: 30px;
		}
		.container .btn {
			background-color: #28a745;
			color: #ffffff;
			padding: 15px 30px;
			text-decoration: none;
			font-size: 1em;
			border-radius: 5px;
			transition: background-color 0.3s ease;
		}
		.container .btn:hover {
			background-color: #218838;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Email Verified Successfully!</h1>
		<p>Your email has been verified. You may now log in to your account.</p>
		<a href="/auth/login" class="btn">Go to Login</a>
	</div>
</body>
</html>`

// failureHTML defines the HTML content for failed operations
const failureHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Operation Failed</title>
    <style>
        body {
            background-color: #f8d7da;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            font-family: Arial, sans-serif;
            margin: 0;
        }
        .content {
            background-color: #ffffff;
            padding: 40px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            border-radius: 8px;
            text-align: center;
        }
        .content h1 {
            color: #721c24;
            font-size: 2em;
            margin-bottom: 20px;
        }
        .content p {
            font-size: 1em;
            color: #333333;
            margin-bottom: 30px;
        }
        .content .btn {
            background-color: #721c24;
            color: #ffffff;
            padding: 15px 30px;
            text-decoration: none;
            font-size: 1em;
            border-radius: 5px;
            transition: background-color 0.3s ease;
        }
        .content .btn:hover {
            background-color: #5a1a1a;
        }
    </style>
</head>
<body>
    <div class="content">
        <h1>Operation Failed!</h1>
        <p>We're sorry, but the operation could not be completed. Please try again later.</p>
        <a href="/auth/request-password-reset" class="btn">Request New Reset Link</a>
    </div>
</body>
</html>`

// successResetHTML defines the HTML content for successful password reset
const successResetHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Password Reset Successful</title>
	<style>
		body {
			background-color: #d4edda;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
			font-family: Arial, sans-serif;
			margin: 0;
		}
		.container {
			background-color: #ffffff;
			padding: 40px;
			box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
			border-radius: 8px;
			text-align: center;
		}
		.container h1 {
			color: #155724;
			font-size: 2em;
			margin-bottom: 20px;
		}
		.container p {
			font-size: 1em;
			color: #333333;
			margin-bottom: 30px;
		}
		.container .btn {
			background-color: #28a745;
			color: #ffffff;
			padding: 15px 30px;
			text-decoration: none;
			font-size: 1em;
			border-radius: 5px;
			transition: background-color 0.3s ease;
		}
		.container .btn:hover {
			background-color: #218838;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Password Reset Successful!</h1>
		<p>Your password has been reset successfully. You may now log in with your new password.</p>
		<a href="/auth/login" class="btn">Go to Login</a>
	</div>
</body>
</html>`
