package handlers

import (
	"encoding/json"
	"fmt"
	"gptimg/internal/config"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"gptimg/pkg/response"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LLMConfigHandler struct {
	repo *repository.LLMConfigRepository
	cfg  *config.Config
}

type CreateLLMConfigRequest struct {
	ConfigName           string `json:"config_name" binding:"required"`
	APIKey               string `json:"api_key"`
	APIBaseURL           string `json:"api_base_url"`
	Model                string `json:"model"`
	IsActive             bool   `json:"is_active"`
	MaxRequestsPerMinute int    `json:"max_requests_per_minute"`
}

func NewLLMConfigHandler(repo *repository.LLMConfigRepository, cfg *config.Config) *LLMConfigHandler {
	return &LLMConfigHandler{repo: repo, cfg: cfg}
}

func (h *LLMConfigHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req CreateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.APIKey == "" {
		response.BadRequest(c, "API key is required")
		return
	}
	if req.Model == "" {
		req.Model = "deepseek-chat"
	}
	if req.MaxRequestsPerMinute == 0 {
		req.MaxRequestsPerMinute = 5
	}
	encrypted, err := utils.EncryptAPIKey(req.APIKey, h.cfg.EncryptionKey)
	if err != nil {
		response.InternalError(c, "Failed to encrypt API key")
		return
	}
	item := &models.LLMConfig{
		UserID:               userID.(int),
		ConfigName:           req.ConfigName,
		APIKeyEncrypted:      encrypted,
		APIBaseURL:           req.APIBaseURL,
		Model:                req.Model,
		IsActive:             req.IsActive,
		MaxRequestsPerMinute: req.MaxRequestsPerMinute,
	}
	if err := h.repo.Create(item); err != nil {
		response.InternalError(c, "Failed to create LLM config")
		return
	}
	response.Success(c, item)
}

func (h *LLMConfigHandler) GetList(c *gin.Context) {
	items, err := h.repo.FindAll()
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	response.Success(c, items)
}

func (h *LLMConfigHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	item, err := h.repo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if item == nil {
		response.NotFound(c, "LLM config not found")
		return
	}
	var req CreateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.APIKey != "" {
		encrypted, err := utils.EncryptAPIKey(req.APIKey, h.cfg.EncryptionKey)
		if err != nil {
			response.InternalError(c, "Failed to encrypt API key")
			return
		}
		item.APIKeyEncrypted = encrypted
	}
	item.ConfigName = req.ConfigName
	item.APIBaseURL = req.APIBaseURL
	if req.Model != "" {
		item.Model = req.Model
	}
	item.IsActive = req.IsActive
	if req.MaxRequestsPerMinute > 0 {
		item.MaxRequestsPerMinute = req.MaxRequestsPerMinute
	}
	if err := h.repo.Update(item); err != nil {
		response.InternalError(c, "Failed to update LLM config")
		return
	}
	response.Success(c, item)
}

func (h *LLMConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.repo.Delete(id); err != nil {
		response.InternalError(c, "Failed to delete LLM config")
		return
	}
	response.Success(c, gin.H{"message": "LLM config deleted"})
}

func (h *LLMConfigHandler) GetPoolStatus(c *gin.Context) {
	items, err := h.repo.FindAll()
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	statuses := make([]*models.LLMConfigStatus, 0, len(items))
	for _, item := range items {
		status := &models.LLMConfigStatus{
			ID:                   item.ID,
			ConfigName:           item.ConfigName,
			APIBaseURL:           item.APIBaseURL,
			Model:                item.Model,
			IsActive:             item.IsActive,
			MaxRequestsPerMinute: item.MaxRequestsPerMinute,
			CheckedAt:            time.Now(),
			StatusMessage:        "inactive",
		}
		if item.IsActive {
			apiKey, err := utils.DecryptAPIKey(item.APIKeyEncrypted, h.cfg.EncryptionKey)
			if err != nil {
				status.StatusMessage = "failed to decrypt key"
			} else {
				healthy, code, message := checkLLMConnection(item.APIBaseURL, apiKey)
				status.Healthy = healthy
				status.StatusCode = code
				status.StatusMessage = message
			}
		}
		statuses = append(statuses, status)
	}
	response.Success(c, statuses)
}

func checkLLMConnection(baseURL, apiKey string) (bool, int, string) {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	req, err := http.NewRequest("GET", strings.TrimRight(baseURL, "/")+"/models", nil)
	if err != nil {
		return false, 0, "invalid LLM URL"
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, resp.StatusCode, "connected"
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = fmt.Sprintf("status %d", resp.StatusCode)
	} else {
		var payload map[string]interface{}
		if json.Unmarshal(body, &payload) == nil {
			if errorObj, ok := payload["error"].(map[string]interface{}); ok {
				if msg, ok := errorObj["message"].(string); ok && msg != "" {
					message = msg
				}
			}
		}
	}
	return false, resp.StatusCode, message
}
