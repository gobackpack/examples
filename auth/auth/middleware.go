package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		claims, valid := extractToken(token)
		if claims == nil || !valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userId := claims["sub"]
		clientId := claims["client_id"]
		// TODO: check if clientId for userId exists in cache
		if userId == nil || clientId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		logrus.Infof("userId -> %v, clientId -> %v ", userId, clientId)

		ctx.Next()
	}
}
