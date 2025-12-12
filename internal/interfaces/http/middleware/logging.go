package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware логирует все HTTP запросы
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Логируем после обработки
		duration := time.Since(startTime)
		fmt.Printf(
			"[%s] %s %s - %d (%dms)\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Milliseconds(),
		)
	}
}

// CORS middleware для локальной разработки
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
