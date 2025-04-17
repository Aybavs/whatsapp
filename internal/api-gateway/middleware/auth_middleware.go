package middleware

import (
	"net/http"
	"strings"

	"whatsapp/pkg/auth"

	"github.com/gin-gonic/gin"
)

// Global auth service instance for the AuthRequired middleware
var authServiceInstance *auth.Service

// SetAuthService sets the global auth service for the AuthRequired middleware
func SetAuthService(service *auth.Service) {
    authServiceInstance = service
}

// AuthRequired middleware verifies user authentication using the globally set auth service
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        if authServiceInstance == nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Auth service not initialized"})
            return
        }
        
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            return
        }

        tokenString := parts[1]
        claims, err := authServiceInstance.ValidateToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            return
        }

        c.Set("UserID", claims.UserID)
        c.Set("username", claims.Username)

        c.Next()
    }
}

// AuthMiddleware creates a middleware that validates JWT tokens using the provided auth service
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            return
        }

        tokenString := parts[1]
        claims, err := authService.ValidateToken(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            return
        }

        c.Set("UserID", claims.UserID)
        c.Set("username", claims.Username)

        c.Next()
    }
}