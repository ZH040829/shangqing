package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"shangqing/internal/service"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix         = "Bearer "
	UserIDKey            = "user_id"
	UsernameKey         = "username"
)

// JWTAuth JWT 认证中间件
func JWTAuth(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		token := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := userSvc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Next()
	}
}

// GetUserID 获取当前用户 ID
func GetUserID(c *gin.Context) int64 {
	if id, exists := c.Get(UserIDKey); exists {
		return id.(int64)
	}
	return 0
}

// GetUsername 获取当前用户名
func GetUsername(c *gin.Context) string {
	if name, exists := c.Get(UsernameKey); exists {
		return name.(string)
	}
	return ""
}

// CORS CORS 中间件
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return "[" + param.TimeStamp.Format("2006-01-02 15:04:05") + "] " +
			param.Method + " " +
			param.Path + " " +
			param.StatusCodeColor() + " " +
			param.Latency.String() + " " +
			param.ClientIP + "\n"
	})
}

// RateLimit 简单限流中间件
func RateLimit(redis *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 使用 Redis 实现滑动窗口限流
		c.Next()
	}
}

type RateLimiter struct{}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}
