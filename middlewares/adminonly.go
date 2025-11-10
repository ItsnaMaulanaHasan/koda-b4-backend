package middlewares

import (
	"backend-daily-greens/lib"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, lib.ResponseError{
				Success: false,
				Message: "Access forbidden: admin only",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
