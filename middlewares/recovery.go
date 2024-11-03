// middlewares/recovery.go
package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
)

// RecoveryMiddleware menangani panic dan mengembalikan respons 500
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                utils.Logger.Errorf("Panic recovered: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "status":  "error",
                    "message": "Internal Server Error",
                })
                c.Abort()
            }
        }()

        c.Next()
    }
}
