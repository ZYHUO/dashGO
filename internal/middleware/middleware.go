package middleware

import (
	"net/http"
	"strings"

	"dashgo/internal/model"
	"dashgo/internal/service"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
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

// JWTAuth JWT 认证中间件
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

		// 使用类型断言检查 IsAdmin 字段
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

// NodeAuth 节点认证中间件
func NodeAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 或 Query 获取 token
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

// RateLimit 速率限制中间件
func RateLimit(limit int) gin.HandlerFunc {
	// TODO: 实现速率限制
	return func(c *gin.Context) {
		c.Next()
	}
}

// SecurityHeaders 安全响应头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 XSS 攻击
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		
		// 严格传输安全（HTTPS）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		c.Next()
	}
}

// InputSanitization 输入清理中间件（防止 XSS）
func InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查常见的恶意模式
		dangerousPatterns := []string{
			"<script", "</script>", "javascript:", "onerror=", "onload=",
			"eval(", "expression(", "vbscript:", "data:text/html",
		}
		
		// 检查 URL 参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				valueLower := strings.ToLower(value)
				for _, pattern := range dangerousPatterns {
					if strings.Contains(valueLower, pattern) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input detected"})
						c.Abort()
						return
					}
				}
			}
			_ = key // 避免未使用变量警告
		}
		
		c.Next()
	}
}

// IPWhitelist IP 白名单中间件（用于管理后台）
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			// 如果没有配置白名单，则允许所有 IP
			c.Next()
			return
		}
		
		clientIP := c.ClientIP()
		allowed := false
		
		for _, ip := range allowedIPs {
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied from this IP"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
