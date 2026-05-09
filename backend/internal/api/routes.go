package api

import (
	"gptimg/internal/api/handlers"
	"gptimg/internal/api/middleware"
	"gptimg/internal/config"
	"gptimg/internal/repository"
	"gptimg/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	router.Static("/storage", cfg.StoragePath)

	userRepo := repository.NewUserRepository()
	imageRepo := repository.NewImageRepository()
	sessionRepo := repository.NewSessionRepository()
	statsRepo := repository.NewStatsRepository()
	configRepo := repository.NewConfigRepository()

	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	imageService := services.NewImageService(configRepo, imageRepo, userRepo, statsRepo, cfg)
	imageHandler := handlers.NewImageHandler(imageService, imageRepo)
	sessionHandler := handlers.NewSessionHandler(sessionRepo, imageRepo)
	statsHandler := handlers.NewStatsHandler(statsRepo)
	configHandler := handlers.NewConfigHandler(configRepo, cfg)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetMe)
		}

		images := v1.Group("/images")
		images.Use(middleware.AuthMiddleware(cfg))
		{
			images.POST("/generate", imageHandler.Generate)
			images.GET("/:id", imageHandler.GetByID)
			images.DELETE("/:id", imageHandler.Delete)
		}

		history := v1.Group("/history")
		history.Use(middleware.AuthMiddleware(cfg))
		{
			history.GET("", imageHandler.GetHistory)
		}

		sessions := v1.Group("/sessions")
		sessions.Use(middleware.AuthMiddleware(cfg))
		{
			sessions.GET("", sessionHandler.GetList)
			sessions.POST("", sessionHandler.Create)
			sessions.GET("/:id", sessionHandler.GetByID)
			sessions.GET("/:id/messages", sessionHandler.GetMessages)
			sessions.DELETE("/:id", sessionHandler.Delete)
		}

		stats := v1.Group("/stats")
		stats.Use(middleware.AuthMiddleware(cfg))
		{
			stats.GET("/overview", statsHandler.GetOverview)
			stats.GET("/daily", statsHandler.GetDaily)
		}

		configGroup := v1.Group("/config")
		configGroup.Use(middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
		{
			configGroup.GET("", configHandler.GetList)
			configGroup.POST("", configHandler.Create)
			configGroup.PUT("/:id", configHandler.Update)
			configGroup.DELETE("/:id", configHandler.Delete)
		}
	}

	return router
}
