package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gobackpack/examples/auth/auth"
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

	public := router.Group("/api")

	public.POST("register", func(ctx *gin.Context) {
		var req *RegisterRequest
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("registration failed: %v", err))
			return
		}

		user := &auth.User{
			Email:    req.Email,
			Password: req.Password,
		}

		if err := auth.RegisterUser(user); err != nil {
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

		token, err := auth.Authenticate(req.Email, req.Password)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, fmt.Sprintf("login failed: %v", err))
			return
		}

		ctx.JSON(http.StatusOK, token)
	})

	protected := router.Group("api/users")
	protected.Use(authRequired())

	protected.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, auth.Users)
	})

	httpServe(router, "", "8080")
}

func authRequired() gin.HandlerFunc {
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

		if isAuthenticated := auth.IsAuthenticated(token); !isAuthenticated {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}
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
