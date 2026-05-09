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
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func normalizeImageSize(size string) string {
	switch strings.TrimSpace(strings.ToLower(size)) {
	case "", "auto":
		return "1024x1024"
	default:
		return size
	}
}

func normalizeImageQuality(quality string) string {
	switch strings.TrimSpace(strings.ToLower(quality)) {
	case "", "auto":
		return "medium"
	case "standard":
		return "medium"
	case "hd":
		return "high"
	default:
		return quality
	}
}

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
	Prompt             string           `json:"prompt" form:"prompt" binding:"required"`
	SessionID          string           `json:"session_id" form:"session_id"`
	Size               string           `json:"size" form:"size"`
	Quality            string           `json:"quality" form:"quality"`
	Style              string           `json:"style" form:"style"`
	ReferenceImageData []byte           `json:"-" form:"-"`
	ReferenceImageName string           `json:"-" form:"-"`
	ReferenceImageType string           `json:"-" form:"-"`
	ReferenceImages    []ReferenceImage `json:"-" form:"-"`
}

type ReferenceImage struct {
	Data        []byte
	Name        string
	ContentType string
}

type ChatGPTImageResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
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
	if strings.EqualFold(req.Quality, "hd") || strings.EqualFold(req.Quality, "high") {
		creditsNeeded = 2
	}

	isAdmin := strings.EqualFold(user.Role, "admin")

	if !isAdmin && user.Credits < creditsNeeded {
		return nil, errors.New("insufficient credits")
	}

	apiConfigs, err := s.configRepo.FindActiveConfigs()
	if err != nil {
		return nil, err
	}
	if len(apiConfigs) == 0 {
		return nil, errors.New("no active API configuration found")
	}

	req.Size = normalizeImageSize(req.Size)
	req.Quality = normalizeImageQuality(req.Quality)

	var (
		imageURL       string
		selectedConfig *models.APIConfig
		lastErr        error
	)

	for _, apiConfig := range apiConfigs {
		apiKey, err := utils.DecryptAPIKey(apiConfig.APIKeyEncrypted, s.cfg.EncryptionKey)
		if err != nil {
			lastErr = fmt.Errorf("%s: failed to decrypt API key", apiConfig.ConfigName)
			continue
		}

		imageURL, _, err = s.callChatGPTAPI(apiConfig.APIBaseURL, apiKey, apiConfig.Model, req)
		if err == nil {
			selectedConfig = apiConfig
			break
		}

		lastErr = fmt.Errorf("%s: %w", apiConfig.ConfigName, err)
	}

	if selectedConfig == nil {
		err = lastErr
	}

	if err != nil {
		record := &models.ImageRecord{
			UserID:       userID,
			SessionID:    req.SessionID,
			Prompt:       req.Prompt,
			Size:         req.Size,
			Quality:      req.Quality,
			Style:        req.Style,
			Model:        apiConfigs[0].Model,
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
		RevisedPrompt:  "",
		ImageURL:       imageURL,
		LocalPath:      localPath,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		Model:          selectedConfig.Model,
		CreditsUsed:    creditsNeeded,
		GenerationTime: generationTime,
		Status:         "success",
	}

	if err := s.imageRepo.Create(record); err != nil {
		return nil, err
	}

	if !isAdmin {
		newCredits := user.Credits - creditsNeeded
		if err := s.userRepo.UpdateCredits(userID, newCredits); err != nil {
			return nil, err
		}
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

	if len(req.ReferenceImages) > 0 || len(req.ReferenceImageData) > 0 {
		return s.callChatGPTImageEditAPI(baseURL, apiKey, model, req)
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

	client := &http.Client{Timeout: 330 * time.Second}
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

func (s *ImageService) callChatGPTImageEditAPI(baseURL, apiKey, model string, req *GenerateImageRequest) (string, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("model", model); err != nil {
		return "", "", err
	}
	if err := writer.WriteField("prompt", req.Prompt); err != nil {
		return "", "", err
	}
	if err := writer.WriteField("n", "1"); err != nil {
		return "", "", err
	}
	if req.Size != "" {
		if err := writer.WriteField("size", req.Size); err != nil {
			return "", "", err
		}
	}
	if req.Quality != "" {
		if err := writer.WriteField("quality", req.Quality); err != nil {
			return "", "", err
		}
	}
	if req.Style != "" {
		if err := writer.WriteField("style", req.Style); err != nil {
			return "", "", err
		}
	}

	referenceImages := req.ReferenceImages
	if len(referenceImages) == 0 && len(req.ReferenceImageData) > 0 {
		referenceImages = []ReferenceImage{{
			Data:        req.ReferenceImageData,
			Name:        req.ReferenceImageName,
			ContentType: req.ReferenceImageType,
		}}
	}

	for index, referenceImage := range referenceImages {
		filename := referenceImage.Name
		if filename == "" {
			filename = fmt.Sprintf("reference-%d.png", index+1)
		}

		part, err := writer.CreateFormFile("image", filename)
		if err != nil {
			return "", "", err
		}
		if _, err := io.Copy(part, bytes.NewReader(referenceImage.Data)); err != nil {
			return "", "", err
		}
	}

	if err := writer.Close(); err != nil {
		return "", "", err
	}

	httpReq, err := http.NewRequest("POST", baseURL+"/v1/images/edits", body)
	if err != nil {
		return "", "", err
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 330 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp ChatGPTImageResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
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

func (s *ImageService) LoadReferenceImage(localPath string) ([]byte, string, string, error) {
	if strings.TrimSpace(localPath) == "" {
		return nil, "", "", errors.New("empty local path")
	}

	absolutePath := filepath.Join(s.cfg.StoragePath, localPath)
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, "", "", err
	}

	filename := filepath.Base(localPath)
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	return data, filename, contentType, nil
}
