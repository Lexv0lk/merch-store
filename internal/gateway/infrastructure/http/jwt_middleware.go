package http

import (
	"net/http"
	"strings"

	"github.com/Lexv0lk/merch-store/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	authHeaderName = "Authorization"
)

func NewAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(authHeaderName)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": "missing authorization header"})
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"errors": "invalid auth header"})
			return
		}

		c.Set(jwt.TokenContextKey, parts[1])
		c.Next()
	}
}
