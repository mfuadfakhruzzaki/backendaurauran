// middlewares/validation.go
package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
)

// ValidationMiddleware mengimplementasikan validasi request secara global
func ValidationMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := c.ShouldBindJSON(&c.Request.Body); err != nil {
            var verrs validator.ValidationErrors
            if errors.As(err, &verrs) {
                // Format validation errors
                errorsMap := make(map[string]string)
                for _, e := range verrs {
                    errorsMap[e.Field()] = e.Tag()
                }
                utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed")
                return
            }
            utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request")
            return
        }
        c.Next()
    }
}
