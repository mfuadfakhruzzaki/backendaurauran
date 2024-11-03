// middlewares/auth.go
package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// AuthMiddleware memeriksa validitas JWT token dan apakah token di-blacklist
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.Logger.Warn("Authorization header missing")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            utils.Logger.Warn("Invalid authorization header format")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenStr := parts[1]
        utils.Logger.Debugf("Received token: %s", tokenStr)

        // Periksa apakah token ada dalam blacklist dengan Type 'jwt_blacklist'
        var blacklistedToken models.Token
        err := models.DB.Where("token = ? AND type = ?", tokenStr, models.TokenTypeJWTBlacklist).First(&blacklistedToken).Error
        if err == nil {
            // Token ada dalam blacklist
            utils.Logger.Warnf("Blacklisted token used: %s", tokenStr)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
            c.Abort()
            return
        } else if err != gorm.ErrRecordNotFound {
            // Error lain selain tidak ditemukannya record
            utils.Logger.Errorf("Error checking token blacklist: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            c.Abort()
            return
        } else {
            utils.Logger.Debug("Token not found in blacklist")
        }

        // Parse dan validasi token
        claims, err := utils.ParseJWT(tokenStr)
        if err != nil {
            utils.Logger.Warnf("Invalid token: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        utils.Logger.Debugf("Token valid for user_id: %d, role: %s", claims.UserID, claims.Role)

        // Simpan informasi user di context
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)

        c.Next()
    }
}
