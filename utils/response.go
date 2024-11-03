// utils/response.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse mengirim respons sukses dengan data
func SuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "data":    data,
    })
}

// CreatedResponse mengirim respons sukses saat membuat resource
func CreatedResponse(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, gin.H{
        "status":  "success",
        "data":    data,
    })
}

// ErrorResponse mengirim respons error dengan pesan
func ErrorResponse(c *gin.Context, statusCode int, message string) {
    c.JSON(statusCode, gin.H{
        "status":  "error",
        "message": message,
    })
}
