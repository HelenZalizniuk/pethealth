package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if client sent a request ID, otherwise generate a new one
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}

		// set request ID in context for further use in handlers
		c.Set("RequestID", rid)

		// propagate request ID in response header so client knows the ID of their transaction
		c.Header("X-Request-ID", rid)

		c.Next()
	}
}
