package middleware

import (
	"Groupchat-Service/internal/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

// RequireRoles creates middleware that checks if the user has any of the required roles
func RequireRoles(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from context
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no claims found in token"})
			c.Abort()
			return
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims format"})
			c.Abort()
			return
		}

		// Extract the role from the 'role' field
		roleStr, ok := mapClaims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no role found in token"})
			c.Abort()
			return
		}

		// Parse the role
		userRole, err := models.ParseRole(roleStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid role in token: %v", err)})
			c.Abort()
			return
		}

		// Check if the user's role matches any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf(
					"insufficient permissions",
				),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func convertRolesToStrings(roles []models.Role) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}
	return result
}
