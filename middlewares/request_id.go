// middlewares/request_id.go
package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware menambahkan Request ID pada setiap request
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Writer.Header().Set("X-Request-ID", requestID)
        c.Next()
    }
}
