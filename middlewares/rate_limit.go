// middlewares/rate_limit.go
package middlewares

import (
	"net/http"
	"time"

	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
    visitors = make(map[string]*rate.Limiter)
    mu       sync.Mutex
    // Limit per IP: misalnya 10 request per second dengan burst 20
    rateLimit = rate.Every(time.Second / 10)
    burst     = 20
)

// getVisitor mengembalikan rate limiter untuk IP tertentu
func getVisitor(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()

    limiter, exists := visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rateLimit, burst)
        visitors[ip] = limiter
    }

    return limiter
}

// RateLimitMiddleware membatasi jumlah request per IP
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter := getVisitor(ip)

        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "status":  "error",
                "message": "Too many requests",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
