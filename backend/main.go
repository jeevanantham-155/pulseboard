package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lords/live-polling/backend/config"
	"github.com/lords/live-polling/backend/controllers"
	"github.com/lords/live-polling/backend/repository"
	"github.com/lords/live-polling/backend/routes"
	"github.com/lords/live-polling/backend/services/auth"
	"github.com/lords/live-polling/backend/services/mongodb"
	"github.com/lords/live-polling/backend/services/poll"
	"github.com/lords/live-polling/backend/services/redis"
	pollwebsocket "github.com/lords/live-polling/backend/websocket"
)

const shutdownTimeout = 10 * time.Second

func main() {
	applicationConfig, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration load failed: %v\n", err)
		os.Exit(1)
	}

	startupContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	mongoClient, err := mongodb.Connect(startupContext, applicationConfig.MongoDBURI, applicationConfig.MongoDBDatabase)
	if err != nil {
		cancel()
		fmt.Fprintf(os.Stderr, "mongodb startup check failed: %v\n", err)
		os.Exit(1)
	}

	redisClient, err := redis.Connect(startupContext, applicationConfig.RedisURL)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		cancel()
		fmt.Fprintf(os.Stderr, "redis startup check failed: %v\n", err)
		os.Exit(1)
	}
	cancel()
	defer closeClients(mongoClient, redisClient)

	userRepository := repository.NewUserRepository(mongoClient.Database())
	indexContext, indexCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	if err := userRepository.EnsureIndexes(indexContext); err != nil {
		indexCancel()
		fmt.Fprintf(os.Stderr, "mongodb user index setup failed: %v\n", err)
		os.Exit(1)
	}
	voteRepository := repository.NewVoteRepository(mongoClient.Database())
	if err := voteRepository.EnsureIndexes(indexContext); err != nil {
		indexCancel()
		fmt.Fprintf(os.Stderr, "mongodb vote index setup failed: %v\n", err)
		os.Exit(1)
	}
	indexCancel()
	authService := auth.NewService(userRepository, applicationConfig.JWTSecret)
	authController := controllers.NewAuthController(authService)
	pollRepository := repository.NewPollRepository(mongoClient.Database())
	pollService := poll.NewService(pollRepository)
	pollController := controllers.NewPollController(pollService)
	voteService := poll.NewVoteService(pollRepository, voteRepository, redisClient)
	publicPollController := controllers.NewPublicPollController(pollService, voteService)
	expiryWorker := poll.NewExpiryWorker(pollService, redisClient)
	expiryWorker.Start()
	defer expiryWorker.Stop()
	websocketManager := pollwebsocket.NewManager(redisClient, applicationConfig.FrontendURL)
	defer websocketManager.Close()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(corsMiddleware(applicationConfig.FrontendURL))
	router.GET("/api/health", healthHandler(mongoClient, redisClient))
	routes.RegisterAuthRoutes(router.Group("/api"), authController)
	routes.RegisterPollRoutes(router.Group("/api"), pollController, authService)
	routes.RegisterPublicRoutes(router.Group("/api"), publicPollController)
	routes.RegisterWebSocketRoutes(router, websocketManager)

	server := &http.Server{
		Addr:              ":" + applicationConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "http server failed: %v\n", err)
		}
	case <-shutdownSignal:
		shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			fmt.Fprintf(os.Stderr, "http server shutdown failed: %v\n", err)
		}
	}
}

func corsMiddleware(frontendURL string) gin.HandlerFunc {
	return func(context *gin.Context) {
		origin := context.GetHeader("Origin")
		if origin != "" && origin != frontendURL {
			context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "origin is not allowed"})
			return
		}
		if origin != "" {
			context.Header("Access-Control-Allow-Origin", origin)
			context.Header("Vary", "Origin")
			context.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Voter-Key")
			context.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if context.Request.Method == http.MethodOptions {
			context.Status(http.StatusNoContent)
			context.Abort()
			return
		}
		context.Next()
	}
}

func closeClients(mongoClient *mongodb.Client, redisClient *redis.Client) {
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := redisClient.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "redis shutdown failed: %v\n", err)
	}
	if err := mongoClient.Disconnect(shutdownContext); err != nil {
		fmt.Fprintf(os.Stderr, "mongodb shutdown failed: %v\n", err)
	}
}

func healthHandler(mongoClient *mongodb.Client, redisClient *redis.Client) gin.HandlerFunc {
	return func(context *gin.Context) {
		mongoStatus := "connected"
		if err := mongoClient.Ping(context.Request.Context()); err != nil {
			mongoStatus = "disconnected"
		}

		redisStatus := "connected"
		if err := redisClient.Ping(context.Request.Context()); err != nil {
			redisStatus = "disconnected"
		}

		status := http.StatusOK
		healthStatus := "ok"
		if mongoStatus != "connected" || redisStatus != "connected" {
			status = http.StatusServiceUnavailable
			healthStatus = "error"
		}

		context.JSON(status, gin.H{
			"status":  healthStatus,
			"mongodb": mongoStatus,
			"redis":   redisStatus,
		})
	}
}
