// middlewares/auth.go
package middlewares

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"

	"github.com/golang-jwt/jwt/v4" // Use maintained JWT library
)

// Context keys as constants to avoid typos
const (
	ContextUserKey = "user"
)

// CustomClaims represents the JWT claims structure
type CustomClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// AuthMiddleware checks JWT token validity and blacklist status
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Logger.Warn("Authorization header missing")
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header missing")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Logger.Warn("Invalid authorization header format")
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		// Avoid logging the entire token for security reasons
		maskedToken := maskToken(tokenStr)
		utils.Logger.Debugf("Received token: %s", maskedToken)

		// Check if token is blacklisted
		var blacklistedToken models.Token
		err := models.DB.Where("token = ? AND type = ?", tokenStr, "jwt_blacklist").First(&blacklistedToken).Error
		if err == nil {
			// Token is blacklisted
			utils.Logger.Warnf("Blacklisted token used: %s", maskedToken)
			utils.ErrorResponse(c, http.StatusUnauthorized, "Token has been revoked")
			c.Abort()
			return
		} else if err != gorm.ErrRecordNotFound {
			// Other database errors
			utils.Logger.Errorf("Error checking token blacklist: %v", err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
			c.Abort()
			return
		} else {
			utils.Logger.Debug("Token not found in blacklist")
		}

		// Parse and validate token
		claims, err := parseJWT(tokenStr)
		if err != nil {
			utils.Logger.Warnf("Invalid token: %v", err)
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		utils.Logger.Debugf("Token valid for user_id: %d, role: %s", claims.UserID, claims.Role)

		// Retrieve full User object from the database
		var user models.User
		if err := models.DB.First(&user, claims.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				utils.Logger.Warnf("User not found for ID: %d", claims.UserID)
				utils.ErrorResponse(c, http.StatusUnauthorized, "User not found")
			} else {
				utils.Logger.Errorf("Error retrieving user: %v", err)
				utils.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
			}
			c.Abort()
			return
		}

		// Ensure user has necessary fields
		if user.Username == "" || user.Email == "" || user.Role == "" {
			utils.Logger.Warnf("User data incomplete for ID: %d", user.ID)
			utils.ErrorResponse(c, http.StatusInternalServerError, "User data incomplete")
			c.Abort()
			return
		}

		// Set the complete User object in context
		c.Set(ContextUserKey, user)

		c.Next()
	}
}

// parseJWT parses and validates the JWT token.
func parseJWT(tokenStr string) (*CustomClaims, error) {
	// Define the claims structure
	claims := &CustomClaims{}

	// Parse the token with the claims
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return nil, errors.New("JWT_SECRET not set in environment")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate the token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Additional claims validation if needed
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// maskToken masks a JWT token for safe logging.
func maskToken(token string) string {
	if len(token) <= 10 {
		return "****"
	}
	return token[:5] + "****" + token[len(token)-5:]
}
