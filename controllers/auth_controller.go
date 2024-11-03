// controllers/auth_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest represents the request structure for user registration
type RegisterRequest struct {
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

	// Check if user already exists
	var existingUser models.User
	if err := models.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email already registered")
		return
	} else if err != gorm.ErrRecordNotFound {
		utils.Logger.Errorf("Failed to check existing user: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Create new user
	user := models.User{
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
    // Ambil token dari header Authorization
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        utils.ErrorResponse(c, http.StatusBadRequest, "Authorization header required")
        return
    }

    // Ekstrak token
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        utils.ErrorResponse(c, http.StatusBadRequest, "Invalid authorization header format")
        return
    }
    tokenStr := parts[1]

    // Parse JWT token untuk mendapatkan klaim
    claims, err := utils.ParseJWT(tokenStr)
    if err != nil {
        utils.Logger.Errorf("Failed to parse JWT token: %v", err)
        utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
        return
    }

    // Tambahkan token ke blacklist dengan UserID yang diambil dari klaim
    blacklistToken := models.Token{
        UserID:    claims.UserID, // Pastikan UserID diisi
        Token:     tokenStr,
        Type:      models.TokenTypeJWTBlacklist,
        ExpiresAt: time.Now().Add(time.Hour * 24), // Sesuaikan durasi blacklist
    }

    if err := models.DB.Create(&blacklistToken).Error; err != nil {
        utils.Logger.Errorf("Failed to blacklist token: %v", err)
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to logout")
        return
    }

    utils.Logger.Infof("Token blacklisted successfully for user ID: %d", claims.UserID)

    // Kirim respons sukses
    utils.SuccessResponse(c, gin.H{
        "message": "Successfully logged out",
    })
}


// VerifyEmail handles email verification
func VerifyEmail(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", tokenStr, models.TokenTypeEmailVerify).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render halaman gagal dengan pesan
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		utils.Logger.Errorf("Failed to verify email token: %v", err)
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render halaman gagal dengan pesan token kedaluwarsa
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var user models.User
	if err := models.DB.First(&user, token.UserID).Error; err != nil {
		utils.Logger.Errorf("User not found for email verification: %v", err)
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	user.IsEmailVerified = true

	if err := models.DB.Save(&user).Error; err != nil {
		utils.Logger.Errorf("Failed to update user email verification: %v", err)
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Delete the token after verification
	if err := models.DB.Delete(&token).Error; err != nil {
		utils.Logger.Errorf("Failed to delete email verification token: %v", err)
		// Tidak mengembalikan error karena email sudah diverifikasi
	}

	utils.Logger.Infof("Email verified successfully for user: %s", user.Email)

	// Render halaman sukses verifikasi email
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
			// Tidak mengungkapkan apakah email ada atau tidak untuk keamanan
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
	// Ambil token dari query parameter
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", tokenStr, models.TokenTypePasswordReset).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render halaman gagal dengan pesan
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		utils.Logger.Errorf("Failed to verify reset token: %v", err)
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render halaman gagal dengan pesan token kedaluwarsa
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Render halaman reset password dengan form
	// Anda bisa mengintegrasikan token dalam form sebagai hidden field
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

	// Replace {{.Token}} dengan token yang sebenarnya
	finalHTML := strings.Replace(htmlContent, "{{.Token}}", tokenStr, 1)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(finalHTML))
}


// ResetPassword handles resetting the user's password using the provided token (POST request)
func ResetPassword(c *gin.Context) {
	var req ResetPasswordFormRequest
	// Bind form data ke struct
	if err := c.ShouldBind(&req); err != nil {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>"+"Invalid input: "+err.Error()+"</p>"))
		return
	}

	// Validasi bahwa password dan konfirmasi password cocok
	if req.NewPassword != req.ConfirmPassword {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte("<p>Passwords do not match</p>"))
		return
	}

	var token models.Token
	if err := models.DB.Where("token = ? AND type = ?", req.Token, models.TokenTypePasswordReset).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Render halaman gagal dengan pesan
			c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
			return
		}
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	if token.ExpiresAt.Before(time.Now()) {
		// Render halaman gagal dengan pesan token kedaluwarsa
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	var user models.User
	if err := models.DB.First(&user, token.UserID).Error; err != nil {
		// Render halaman gagal dengan pesan generik
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Hash password secara manual
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Update password dengan hash baru
	user.Password = string(hashedPassword)

	if err := models.DB.Save(&user).Error; err != nil {
		// Render halaman gagal dengan pesan
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(failureHTML))
		return
	}

	// Delete the token after use
	if err := models.DB.Delete(&token).Error; err != nil {
		// Tidak mengembalikan error karena password sudah diubah
	}

	// Render halaman sukses reset password
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
			background-color: #d4edda; /* Latar belakang hijau muda */
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh; /* Tinggi penuh viewport */
			font-family: Arial, sans-serif;
			margin: 0;
		}
		.container {
			background-color: #ffffff; /* Latar belakang putih */
			padding: 40px;
			box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); /* Bayangan lembut */
			border-radius: 8px; /* Sudut membulat */
			text-align: center;
		}
		.container h1 {
			color: #155724; /* Warna hijau untuk judul */
			font-size: 2em;
			margin-bottom: 20px;
		}
		.container p {
			font-size: 1em;
			color: #333333;
			margin-bottom: 30px;
		}
		.container .btn {
			background-color: #28a745; /* Tombol hijau */
			color: #ffffff;
			padding: 15px 30px;
			text-decoration: none;
			font-size: 1em;
			border-radius: 5px;
			transition: background-color 0.3s ease;
		}
		.container .btn:hover {
			background-color: #218838; /* Warna hijau lebih gelap saat di-hover */
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
    <title>Password Reset Failed</title>
    <style>
        body {
            background-color: #f8d7da; /* Latar belakang merah muda */
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh; /* Tinggi penuh viewport */
            font-family: Arial, sans-serif;
            margin: 0;
        }
        .content {
            background-color: #ffffff; /* Latar belakang putih */
            padding: 40px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); /* Bayangan lembut */
            border-radius: 8px; /* Sudut membulat */
            text-align: center;
        }
        .content h1 {
            color: #721c24; /* Warna merah untuk judul */
            font-size: 2em;
            margin-bottom: 20px;
        }
        .content p {
            font-size: 1em;
            color: #333333;
            margin-bottom: 30px;
        }
        .content .btn {
            background-color: #721c24; /* Tombol merah */
            color: #ffffff;
            padding: 15px 30px;
            text-decoration: none;
            font-size: 1em;
            border-radius: 5px;
            transition: background-color 0.3s ease;
        }
        .content .btn:hover {
            background-color: #5a1a1a; /* Warna merah lebih gelap saat di-hover */
        }
    </style>
</head>
<body>
    <div class="content">
        <h1>Password Reset Failed!</h1>
        <p>We're sorry, but the password reset link is invalid or has expired. Please request a new one.</p>
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
			background-color: #d4edda; /* Latar belakang hijau muda */
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh; /* Tinggi penuh viewport */
			font-family: Arial, sans-serif;
			margin: 0;
		}
		.container {
			background-color: #ffffff; /* Latar belakang putih */
			padding: 40px;
			box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); /* Bayangan lembut */
			border-radius: 8px; /* Sudut membulat */
			text-align: center;
		}
		.container h1 {
			color: #155724; /* Warna hijau untuk judul */
			font-size: 2em;
			margin-bottom: 20px;
		}
		.container p {
			font-size: 1em;
			color: #333333;
			margin-bottom: 30px;
		}
		.container .btn {
			background-color: #28a745; /* Tombol hijau */
			color: #ffffff;
			padding: 15px 30px;
			text-decoration: none;
			font-size: 1em;
			border-radius: 5px;
			transition: background-color 0.3s ease;
		}
		.container .btn:hover {
			background-color: #218838; /* Warna hijau lebih gelap saat di-hover */
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


