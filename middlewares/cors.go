// middlewares/cors.go
package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware mengatur CORS untuk aplikasi
func CORSMiddleware() gin.HandlerFunc {
    configCORS := cors.Config{
        AllowOrigins:     []string{"http://localhost:5173, https://zacht.tech"}, // Ganti dengan domain yang diizinkan di produksi
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }

    return cors.New(configCORS)
}
