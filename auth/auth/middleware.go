package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (authSvc *Service) RequiredAuthentication() gin.HandlerFunc {
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
		accessTokenUuid := claims["uuid"]
		if userId == nil || accessTokenUuid == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		bUserId, err := authSvc.Cache.Get(fmt.Sprint(accessTokenUuid))
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if fmt.Sprint(userId) != string(bUserId) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Next()
	}
}
