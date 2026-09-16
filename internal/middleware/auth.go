package middleware

import (
	"net/http"
	"strings"

	"ticket-system/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "user_id"
)

// AuthMiddleware validates the JWT from Authorization header and sets user_id in context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header format must be Bearer <token>"})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from Gin context.
func GetUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, false
	}
	userID, ok := val.(int64)
	return userID, ok
}
