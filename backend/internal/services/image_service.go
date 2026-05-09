package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gptimg/internal/config"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/internal/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageService struct {
	configRepo *repository.ConfigRepository
	imageRepo  *repository.ImageRepository
	userRepo   *repository.UserRepository
	statsRepo  *repository.StatsRepository
	cfg        *config.Config
}

func NewImageService(
	configRepo *repository.ConfigRepository,
	imageRepo *repository.ImageRepository,
	userRepo *repository.UserRepository,
	statsRepo *repository.StatsRepository,
	cfg *config.Config,
) *ImageService {
	return &ImageService{
		configRepo: configRepo,
		imageRepo:  imageRepo,
		userRepo:   userRepo,
		statsRepo:  statsRepo,
		cfg:        cfg,
	}
}

type GenerateImageRequest struct {
	Prompt    string `json:"prompt" binding:"required"`
	SessionID string `json:"session_id"`
	Size      string `json:"size"`
	Quality   string `json:"quality"`
	Style     string `json:"style"`
}

type ChatGPTImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL            string `json:"url"`
		B64JSON        string `json:"b64_json"`
		RevisedPrompt  string `json:"revised_prompt"`
	} `json:"data"`
}

func (s *ImageService) GenerateImage(userID int, req *GenerateImageRequest) (*models.ImageRecord, error) {
	startTime := time.Now()

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	creditsNeeded := 1
	if req.Quality == "hd" {
		creditsNeeded = 2
	}

	if user.Credits < creditsNeeded {
		return nil, errors.New("insufficient credits")
	}

	apiConfig, err := s.configRepo.FindActive()
	if err != nil {
		return nil, err
	}
	if apiConfig == nil {
		return nil, errors.New("no active API configuration found")
	}

	apiKey, err := utils.DecryptAPIKey(apiConfig.APIKeyEncrypted, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	if req.Size == "" {
		req.Size = "1024x1024"
	}
	if req.Quality == "" {
		req.Quality = "standard"
	}

	imageURL, revisedPrompt, err := s.callChatGPTAPI(apiConfig.APIBaseURL, apiKey, apiConfig.Model, req)
	if err != nil {
		record := &models.ImageRecord{
			UserID:       userID,
			SessionID:    req.SessionID,
			Prompt:       req.Prompt,
			Size:         req.Size,
			Quality:      req.Quality,
			Style:        req.Style,
			Model:        apiConfig.Model,
			CreditsUsed:  0,
			Status:       "failed",
			ErrorMessage: err.Error(),
		}
		s.imageRepo.Create(record)

		today := time.Now().Format("2006-01-02")
		s.statsRepo.UpsertDailyStats(userID, today, 0, 0, false)

		return nil, err
	}

	var localPath string
	if strings.HasPrefix(imageURL, "b64:") {
		b64Data := strings.TrimPrefix(imageURL, "b64:")
		localPath, err = s.saveBase64Image(b64Data, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		imageURL = ""
	} else {
		localPath, err = s.downloadAndSaveImage(imageURL, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
	}

	generationTime := int(time.Since(startTime).Milliseconds())

	record := &models.ImageRecord{
		UserID:         userID,
		SessionID:      req.SessionID,
		Prompt:         req.Prompt,
		RevisedPrompt:  revisedPrompt,
		ImageURL:       imageURL,
		LocalPath:      localPath,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		Model:          apiConfig.Model,
		CreditsUsed:    creditsNeeded,
		GenerationTime: generationTime,
		Status:         "success",
	}

	if err := s.imageRepo.Create(record); err != nil {
		return nil, err
	}

	newCredits := user.Credits - creditsNeeded
	if err := s.userRepo.UpdateCredits(userID, newCredits); err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	s.statsRepo.UpsertDailyStats(userID, today, creditsNeeded, generationTime, true)

	return record, nil
}

func (s *ImageService) callChatGPTAPI(baseURL, apiKey, model string, req *GenerateImageRequest) (string, string, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if model == "" {
		model = "gpt-image-2"
	}

	requestBody := map[string]interface{}{
		"model":  model,
		"prompt": req.Prompt,
		"n":      1,
		"size":   req.Size,
	}
	if req.Quality != "" {
		requestBody["quality"] = req.Quality
	}
	if req.Style != "" {
		requestBody["style"] = req.Style
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", err
	}

	httpReq, err := http.NewRequest("POST", baseURL+"/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp ChatGPTImageResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", "", err
	}

	if len(apiResp.Data) == 0 {
		return "", "", errors.New("no image generated")
	}

	item := apiResp.Data[0]

	if item.B64JSON != "" {
		return "b64:" + item.B64JSON, item.RevisedPrompt, nil
	}

	if item.URL != "" {
		return item.URL, item.RevisedPrompt, nil
	}

	return "", "", errors.New("API returned no image URL or base64 data")
}

func (s *ImageService) saveBase64Image(b64Data string, userID int) (string, error) {
	imgData, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return "", err
	}

	now := time.Now()
	dir := filepath.Join(s.cfg.StoragePath, fmt.Sprintf("%d", userID), now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%d.png", now.UnixNano(), userID)
	filePath := filepath.Join(dir, filename)

	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return "", err
	}

	return filepath.Join(fmt.Sprintf("%d", userID), now.Format("2006"), now.Format("01"), filename), nil
}

func (s *ImageService) downloadAndSaveImage(imageURL string, userID int) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	now := time.Now()
	dir := filepath.Join(s.cfg.StoragePath, fmt.Sprintf("%d", userID), now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%d.png", now.Unix(), userID)
	filePath := filepath.Join(dir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	relativePath := filepath.Join(fmt.Sprintf("%d", userID), now.Format("2006"), now.Format("01"), filename)
	return relativePath, nil
}
