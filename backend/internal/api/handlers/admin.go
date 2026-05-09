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

type AdminHandler struct {
	userRepo   *repository.UserRepository
	configRepo *repository.ConfigRepository
	statsRepo  *repository.StatsRepository
	cfg        *config.Config
}

type CreateUserRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	QuotaLimit int    `json:"quota_limit"`
	Role       string `json:"role"`
}

type UpdateUserRequest struct {
	Credits    *int   `json:"credits"`
	QuotaLimit *int   `json:"quota_limit"`
	Status     string `json:"status"`
	Role       string `json:"role"`
}

func NewAdminHandler(userRepo *repository.UserRepository, configRepo *repository.ConfigRepository, statsRepo *repository.StatsRepository, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		userRepo:   userRepo,
		configRepo: configRepo,
		statsRepo:  statsRepo,
		cfg:        cfg,
	}
}

func (h *AdminHandler) GetOverview(c *gin.Context) {
	overview, err := h.statsRepo.GetSystemOverview()
	if err != nil {
		response.InternalError(c, "Failed to load overview")
		return
	}

	configStatuses, err := h.buildConfigStatuses()
	if err != nil {
		response.InternalError(c, "Failed to inspect API pool")
		return
	}

	for _, item := range configStatuses {
		if item.IsActive {
			overview.ActiveAPIConfigs++
		}
		if item.IsActive && item.Healthy {
			overview.HealthyAPIConfigs++
		}
	}

	response.Success(c, overview)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	users, err := h.userRepo.ListWithUsage()
	if err != nil {
		response.InternalError(c, "Failed to load users")
		return
	}

	response.Success(c, users)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if existingUser, err := h.userRepo.FindByEmail(req.Email); err != nil {
		response.InternalError(c, "Database error")
		return
	} else if existingUser != nil {
		response.BadRequest(c, "Email already registered")
		return
	}

	if existingUser, err := h.userRepo.FindByUsername(req.Username); err != nil {
		response.InternalError(c, "Database error")
		return
	} else if existingUser != nil {
		response.BadRequest(c, "Username already exists")
		return
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		response.InternalError(c, "Failed to hash password")
		return
	}

	quotaLimit := req.QuotaLimit
	if quotaLimit <= 0 {
		quotaLimit = 50
	}

	role := req.Role
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "admin" {
		response.BadRequest(c, "Invalid role")
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Credits:      quotaLimit,
		QuotaLimit:   quotaLimit,
		Role:         role,
		Status:       "active",
	}

	if err := h.userRepo.Create(user); err != nil {
		response.InternalError(c, "Failed to create user")
		return
	}

	response.Success(c, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	currentUser, err := h.userRepo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if currentUser == nil {
		response.NotFound(c, "User not found")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	credits := currentUser.Credits
	quotaLimit := currentUser.QuotaLimit
	role := currentUser.Role
	status := currentUser.Status

	if req.Credits != nil {
		credits = *req.Credits
	}
	if req.QuotaLimit != nil {
		quotaLimit = *req.QuotaLimit
	}
	if req.Role != "" {
		if req.Role != "user" && req.Role != "admin" {
			response.BadRequest(c, "Invalid role")
			return
		}
		role = req.Role
	}
	if req.Status != "" {
		if req.Status != "active" && req.Status != "suspended" {
			response.BadRequest(c, "Invalid status")
			return
		}
		status = req.Status
	}

	if credits < 0 || quotaLimit <= 0 {
		response.BadRequest(c, "Invalid credit or quota value")
		return
	}
	if credits > quotaLimit {
		credits = quotaLimit
	}

	if err := h.userRepo.UpdateAdminFields(id, credits, quotaLimit, role, status); err != nil {
		response.InternalError(c, "Failed to update user")
		return
	}

	updatedUser, err := h.userRepo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, updatedUser)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	currentUser, err := h.userRepo.FindByID(id)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if currentUser == nil {
		response.NotFound(c, "User not found")
		return
	}
	if currentUser.Role == "admin" {
		response.Forbidden(c, "Admin accounts cannot be deleted here")
		return
	}

	if err := h.userRepo.DeleteUserWithData(id); err != nil {
		response.InternalError(c, "Failed to delete user")
		return
	}

	response.Success(c, gin.H{"message": "User deleted"})
}

func (h *AdminHandler) GetAPIPoolStatus(c *gin.Context) {
	statuses, err := h.buildConfigStatuses()
	if err != nil {
		response.InternalError(c, "Failed to inspect API pool")
		return
	}

	response.Success(c, statuses)
}

func (h *AdminHandler) buildConfigStatuses() ([]*models.APIConfigStatus, error) {
	configs, err := h.configRepo.FindAll()
	if err != nil {
		return nil, err
	}

	statuses := make([]*models.APIConfigStatus, 0, len(configs))
	for _, cfgItem := range configs {
		status := &models.APIConfigStatus{
			ID:                   cfgItem.ID,
			ConfigName:           cfgItem.ConfigName,
			APIBaseURL:           cfgItem.APIBaseURL,
			Model:                cfgItem.Model,
			IsActive:             cfgItem.IsActive,
			MaxRequestsPerMinute: cfgItem.MaxRequestsPerMinute,
			CheckedAt:            time.Now(),
			StatusMessage:        "inactive",
		}

		if cfgItem.IsActive {
			apiKey, err := utils.DecryptAPIKey(cfgItem.APIKeyEncrypted, h.cfg.EncryptionKey)
			if err != nil {
				status.StatusMessage = "failed to decrypt key"
			} else {
				healthy, statusCode, statusMessage := checkAPIConnection(cfgItem.APIBaseURL, apiKey)
				status.Healthy = healthy
				status.StatusCode = statusCode
				status.StatusMessage = statusMessage
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func checkAPIConnection(baseURL, apiKey string) (bool, int, string) {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	req, err := http.NewRequest("GET", strings.TrimRight(baseURL, "/")+"/v1/models", nil)
	if err != nil {
		return false, 0, "invalid API URL"
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, resp.StatusCode, "connected"
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
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
