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

	public := router.Group("/api")

	public.POST("register", func(ctx *gin.Context) {
		var req *RegisterRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("registration failed: %v", err))
			return
		}

		user, err := authSvc.RegisterUser(req.Email, req.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("registration failed: %v", err))
			return
		}

		ctx.JSON(http.StatusOK, user)
	})

	public.POST("login", func(ctx *gin.Context) {
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

	protected := router.Group("api/users")
	protected.Use(authSvc.RequiredAuthentication())

	protected.POST("/logout", func(ctx *gin.Context) {
		// TODO: Provide implementation
	})

	protected.POST("/token/refresh", func(ctx *gin.Context) {
		// TODO: Provide implementation
	})

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
