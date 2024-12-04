// middlewares/authorization.go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
)

// RoleMiddleware memeriksa apakah pengguna memiliki salah satu peran yang diizinkan
func RoleMiddleware(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil peran pengguna dari konteks
		roleInterface, exists := c.Get("user_role")
		if !exists {
			utils.Logger.Warn("User role not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		userRole, ok := roleInterface.(models.Role)
		if !ok {
			utils.Logger.Warn("User role has invalid type")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role"})
			c.Abort()
			return
		}

		// Periksa apakah peran pengguna termasuk dalam peran yang diizinkan
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		utils.Logger.Warnf("User role '%s' not authorized", userRole)
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		c.Abort()
	}
}
