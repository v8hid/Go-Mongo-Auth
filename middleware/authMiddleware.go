package middleware

import (
	"fmt"
	"net/http"

	helper "v8hid/Go-Mongo-Auth/helpers"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("No authrization header provided")})
			c.Abort()
			return
		}
		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token provided"})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()

	}
}
