package middleware

import (
	"agenda-api/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRoles(models.RoleAdmin)
}

func RequireUser() gin.HandlerFunc {
	return RequireRoles(models.RoleUser, models.RoleAdmin)
}
