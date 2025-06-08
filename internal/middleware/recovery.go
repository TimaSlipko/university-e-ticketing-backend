package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Internal server error",
			"error":   "Something went wrong",
		})
	})
}
