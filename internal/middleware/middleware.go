package middleware

import (
	"net/http"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间�?
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Node-ID, X-Node-Type, X-Node-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, ETag, subscription-userinfo, profile-update-interval, profile-title")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// JWTAuth JWT 认证中间�?
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		user, err := authService.GetUserFromToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// AdminAuth 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// 使用类型断言检�?IsAdmin 字段
		type userWithAdmin interface {
			GetIsAdmin() bool
		}

		// 直接检查结构体字段
		if u, ok := user.(*model.User); ok {
			if !u.IsAdmin {
				c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// NodeAuth 节点认证中间�?
func NodeAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// �?Header �?Query 获取 token
		nodeToken := c.GetHeader("X-Node-Token")
		if nodeToken == "" {
			nodeToken = c.Query("token")
		}

		if nodeToken == "" || nodeToken != token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid node token"})
			c.Abort()
			return
		}

		// 获取节点信息
		nodeID := c.GetHeader("X-Node-ID")
		if nodeID == "" {
			nodeID = c.Query("node_id")
		}

		nodeType := c.GetHeader("X-Node-Type")
		if nodeType == "" {
			nodeType = c.Query("node_type")
		}

		c.Set("node_id", nodeID)
		c.Set("node_type", nodeType)

		c.Next()
	}
}

// RateLimit 速率限制中间�?
func RateLimit(limit int) gin.HandlerFunc {
	// TODO: 实现速率限制
	return func(c *gin.Context) {
		c.Next()
	}
}
