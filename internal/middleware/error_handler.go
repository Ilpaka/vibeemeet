package middleware

import (
	"github.com/gin-gonic/gin"
	"video_conference/pkg/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Проверяем есть ли ошибки
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Определяем статус код
			statusCode := errors.HTTPStatusFromError(err.Err)
			
			c.JSON(statusCode, gin.H{
				"error": err.Error(),
			})
		}
	}
}

