package handlers

import (
	"gptimg/internal/config"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"gptimg/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string        `json:"token"`
	User  *models.User  `json:"user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	response.Forbidden(c, "Accounts are issued by the administrator")
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if user == nil {
		response.Unauthorized(c, "Invalid email or password")
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		response.Unauthorized(c, "Invalid email or password")
		return
	}

	if user.Status != "active" {
		response.Forbidden(c, "Account is suspended")
		return
	}

	token, err := utils.GenerateToken(user, h.cfg.JWTSecret)
	if err != nil {
		response.InternalError(c, "Failed to generate token")
		return
	}

	response.Success(c, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.userRepo.FindByID(userID.(int))
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if user == nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}
