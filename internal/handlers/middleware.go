package handlers

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SetupMiddleware 设置中间件
func SetupMiddleware(r *gin.Engine) {
	// CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 日志中间件
	r.Use(LoggerMiddleware())

	// 错误处理中间件
	r.Use(ErrorHandlerMiddleware())

	// 恢复中间件
	r.Use(gin.Recovery())
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"status":     param.StatusCode,
			"method":     param.Method,
			"path":       param.Path,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
			"latency":    param.Latency,
		}).Info("HTTP Request")
		return ""
	})
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
				"method": c.Request.Method,
			}).Error("Request Error")

			// 如果还没有响应，返回错误响应
			if !c.Writer.Written() {
				InternalErrorResponse(c, "Internal Server Error")
			}
		}
	}
}
