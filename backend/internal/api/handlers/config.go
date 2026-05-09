package handlers

import (
	"gptimg/internal/config"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"gptimg/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configRepo *repository.ConfigRepository
	cfg        *config.Config
}

func NewConfigHandler(configRepo *repository.ConfigRepository, cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
		cfg:        cfg,
	}
}

type CreateConfigRequest struct {
	ConfigName           string `json:"config_name" binding:"required"`
	APIKey               string `json:"api_key"`
	APIBaseURL           string `json:"api_base_url"`
	Model                string `json:"model"`
	IsActive             bool   `json:"is_active"`
	MaxRequestsPerMinute int    `json:"max_requests_per_minute"`
}

func (h *ConfigHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.APIKey == "" {
		response.BadRequest(c, "API key is required")
		return
	}

	encryptedKey, err := utils.EncryptAPIKey(req.APIKey, h.cfg.EncryptionKey)
	if err != nil {
		response.InternalError(c, "Failed to encrypt API key")
		return
	}

	if req.Model == "" {
		req.Model = "gpt-image-2"
	}
	if req.MaxRequestsPerMinute == 0 {
		req.MaxRequestsPerMinute = 5
	}

	apiConfig := &models.APIConfig{
		UserID:               userID.(int),
		ConfigName:           req.ConfigName,
		APIKeyEncrypted:      encryptedKey,
		APIBaseURL:           req.APIBaseURL,
		Model:                req.Model,
		IsActive:             req.IsActive,
		MaxRequestsPerMinute: req.MaxRequestsPerMinute,
	}

	if err := h.configRepo.Create(apiConfig); err != nil {
		response.InternalError(c, "Failed to create config")
		return
	}

	response.Success(c, apiConfig)
}

func (h *ConfigHandler) GetList(c *gin.Context) {
	configs, err := h.configRepo.FindAll()
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, configs)
}

func (h *ConfigHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	config, err := h.configRepo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if config == nil {
		response.NotFound(c, "Config not found")
		return
	}

	if req.APIKey != "" {
		encryptedKey, err := utils.EncryptAPIKey(req.APIKey, h.cfg.EncryptionKey)
		if err != nil {
			response.InternalError(c, "Failed to encrypt API key")
			return
		}
		config.APIKeyEncrypted = encryptedKey
	}

	config.ConfigName = req.ConfigName
	config.APIBaseURL = req.APIBaseURL
	if req.Model != "" {
		config.Model = req.Model
	}
	config.IsActive = req.IsActive
	if req.MaxRequestsPerMinute > 0 {
		config.MaxRequestsPerMinute = req.MaxRequestsPerMinute
	}

	if err := h.configRepo.Update(config); err != nil {
		response.InternalError(c, "Failed to update config")
		return
	}

	response.Success(c, config)
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	if err := h.configRepo.Delete(id); err != nil {
		response.InternalError(c, "Failed to delete config")
		return
	}

	response.Success(c, gin.H{"message": "Config deleted successfully"})
}
