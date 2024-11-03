// middlewares/logging.go
package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
)

// LoggingMiddleware mencatat setiap request yang masuk
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()

        // Lanjutkan ke handler berikutnya
        c.Next()

        // Hitung durasi request
        duration := time.Since(startTime)

        // Ambil status code, method, path, dan IP
        statusCode := c.Writer.Status()
        method := c.Request.Method
        path := c.Request.URL.Path
        clientIP := c.ClientIP()

        // Log informasi request
        utils.Logger.WithFields(map[string]interface{}{
            "status_code": statusCode,
            "method":      method,
            "path":        path,
            "client_ip":   clientIP,
            "duration":    duration,
        }).Info("Handled request")
    }
}
