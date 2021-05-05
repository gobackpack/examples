package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func RequiredAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := strings.Split(ctx.GetHeader("Authorization"), " ")
		if len(authHeader) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		schema, token := authHeader[0], authHeader[1]
		if schema != "Bearer" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !isTokenValid(token) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Next()
	}
}
