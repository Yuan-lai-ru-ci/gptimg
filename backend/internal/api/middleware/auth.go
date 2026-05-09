package middleware

import (
	"strings"

	"gptimg/internal/config"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"gptimg/pkg/response"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	userRepo := repository.NewUserRepository()

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1], cfg.JWTSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		role := claims.Role
		username := claims.Username

		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			response.Unauthorized(c, "Failed to load user")
			c.Abort()
			return
		}
		if user == nil || user.Status != "active" {
			response.Unauthorized(c, "User not available")
			c.Abort()
			return
		}

		role = user.Role
		username = user.Username

		c.Set("user_id", user.ID)
		c.Set("username", username)
		c.Set("role", role)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
