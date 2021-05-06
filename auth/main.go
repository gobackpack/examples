package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gobackpack/examples/auth/auth"
	"github.com/gobackpack/examples/auth/auth/cache/redis"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type RegisterRequest struct {
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	authSvc := &auth.Service{
		Cache: initCacheRepo(15),
	}

	api := router.Group("/api")

	api.POST("register", func(ctx *gin.Context) {
		var req *RegisterRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("registration failed: %v", err))
			return
		}

		user, err := authSvc.RegisterUser(&auth.User{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("registration failed: %v", err))
			return
		}

		ctx.JSON(http.StatusOK, user)
	})

	api.POST("login", func(ctx *gin.Context) {
		var req *LoginRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("login failed: %v", err))
			return
		}

		tokens, err := authSvc.Authenticate(req.Email, req.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("login failed: %v", err))
			return
		}

		ctx.JSON(http.StatusOK, tokens)
	})

	api.POST("logout", authSvc.RequiredAuthentication(), func(ctx *gin.Context) {
		accessToken := auth.GetAccessTokenFromRequest(ctx)

		if err := authSvc.DestroyAuthenticationSession(accessToken); err != nil {
			ctx.JSON(http.StatusBadRequest, "invalid access_token")
			return
		}
	})

	api.POST("token/refresh", authSvc.RequiredAuthentication(), func(ctx *gin.Context) {
		mapToken := map[string]string{}
		if err := ctx.ShouldBindJSON(&mapToken); err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, err.Error())
			return
		}
		refreshToken := mapToken["refresh_token"]

		tokens, err := authSvc.RefreshToken(refreshToken)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("refresh token failed: %v", err))
			return
		}

		ctx.JSON(http.StatusOK, tokens)
	})

	protected := api.Group("users")
	protected.Use(authSvc.RequiredAuthentication())

	protected.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, auth.Users)
	})

	httpServe(router, "", "8080")
}

func httpServe(router *gin.Engine, host, port string) {
	addr := host + ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		logrus.Info("http listen: ", addr)

		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logrus.Error("server listen err: ", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Warn("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("server forced to shutdown: ", err)
	}

	logrus.Warn("server exited")
}

func initCacheRepo(cacheDb int) auth.Cache {
	redisConfig := redis.NewConfig()

	if strings.TrimSpace(redisConfig.Host) == "" {
		redisConfig.Host = ""
	}

	if strings.TrimSpace(redisConfig.Port) == "" {
		redisConfig.Port = "6379"
	}

	if strings.TrimSpace(redisConfig.Password) == "" {
		redisConfig.Password = ""
	}

	redisConfig.DB = cacheDb

	redisConn := &redis.Connection{
		Config: redisConfig,
	}

	if err := redisConn.Initialize(); err != nil {
		logrus.Fatal("failed to initialize redis connection: ", err)
	}

	return redisConn
}
