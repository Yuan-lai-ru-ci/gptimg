package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	llmConfigRepo := repository.NewLLMConfigRepository()

	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	imageService := services.NewImageService(configRepo, imageRepo, userRepo, statsRepo, cfg)
	llmService := services.NewLLMService(llmConfigRepo, cfg)
	imageHandler := handlers.NewImageHandler(imageService, imageRepo)
	sessionHandler := handlers.NewSessionHandler(sessionRepo, imageRepo)
	statsHandler := handlers.NewStatsHandler(statsRepo)
	configHandler := handlers.NewConfigHandler(configRepo, cfg)
	llmConfigHandler := handlers.NewLLMConfigHandler(llmConfigRepo, cfg)
	pptHandler := handlers.NewPPTHandler(llmService, imageService, sessionRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, configRepo, statsRepo, cfg)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

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
			sessions.PUT("/:id", sessionHandler.Update)
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

		llmConfigGroup := v1.Group("/llm-config")
		llmConfigGroup.Use(middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
		{
			llmConfigGroup.GET("", llmConfigHandler.GetList)
			llmConfigGroup.POST("", llmConfigHandler.Create)
			llmConfigGroup.PUT("/:id", llmConfigHandler.Update)
			llmConfigGroup.DELETE("/:id", llmConfigHandler.Delete)
		}

		ppt := v1.Group("/ppt")
		ppt.Use(middleware.AuthMiddleware(cfg))
		{
			ppt.POST("/plan", pptHandler.Plan)
			ppt.POST("/plan-document", pptHandler.PlanDocument)
			ppt.POST("/generate", pptHandler.Generate)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(cfg), middleware.AdminMiddleware())
		{
			admin.GET("/overview", adminHandler.GetOverview)
			admin.GET("/users", adminHandler.GetUsers)
			admin.POST("/users", adminHandler.CreateUser)
			admin.PATCH("/users/:id", adminHandler.UpdateUser)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.GET("/api-pool", adminHandler.GetAPIPoolStatus)
			admin.GET("/llm-pool", llmConfigHandler.GetPoolStatus)
		}
	}

	registerFrontendRoutes(router, cfg)

	return router
}

func registerFrontendRoutes(router *gin.Engine, cfg *config.Config) {
	indexPath := filepath.Join(cfg.FrontendPath, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return
	}

	router.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api/") || strings.HasPrefix(requestPath, "/storage/") {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}

		resolvedPath := resolveFrontendPath(cfg.FrontendPath, requestPath)
		if resolvedPath == "" {
			c.File(indexPath)
			return
		}

		c.File(resolvedPath)
	})
}

func resolveFrontendPath(frontendRoot, requestPath string) string {
	cleanPath := path.Clean("/" + requestPath)
	trimmedPath := strings.TrimPrefix(cleanPath, "/")
	candidates := []string{}

	if trimmedPath == "" || trimmedPath == "." {
		candidates = append(candidates, filepath.Join(frontendRoot, "index.html"))
	} else {
		candidates = append(candidates,
			filepath.Join(frontendRoot, trimmedPath),
			filepath.Join(frontendRoot, trimmedPath+".html"),
			filepath.Join(frontendRoot, trimmedPath, "index.html"),
		)
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate
		}
	}

	return ""
}
