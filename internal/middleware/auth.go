package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"llmgate/internal/auth"
)

const ContextKeyUser = "currentUser"

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwtManager.Validate(parts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ContextKeyUser, claims)
		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户
func GetCurrentUser(c *gin.Context) *auth.Claims {
	user, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil
	}
	return user.(*auth.Claims)
}

// AdminRequired 管理员权限检查
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		c.Next()
	}
}

// ManagerOrAdminRequired 管理员或经理权限检查
func ManagerOrAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if user.Role != "admin" && user.Role != "manager" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "manager or admin access required"})
			return
		}

		c.Next()
	}
}
