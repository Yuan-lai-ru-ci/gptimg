package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    string    `json:"avatar_url"`
	Credits      int       `json:"credits"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID            string    `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	LastMessageAt time.Time `json:"last_message_at"`
	MessageCount  int       `json:"message_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ImageRecord struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	SessionID       string    `json:"session_id"`
	Prompt          string    `json:"prompt"`
	RevisedPrompt   string    `json:"revised_prompt"`
	ImageURL        string    `json:"image_url"`
	LocalPath       string    `json:"local_path"`
	Size            string    `json:"size"`
	Quality         string    `json:"quality"`
	Style           string    `json:"style"`
	Model           string    `json:"model"`
	CreditsUsed     int       `json:"credits_used"`
	GenerationTime  int       `json:"generation_time"`
	Status          string    `json:"status"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

type UsageStats struct {
	ID                     int       `json:"id"`
	UserID                 int       `json:"user_id"`
	Date                   string    `json:"date"`
	TotalGenerations       int       `json:"total_generations"`
	SuccessfulGenerations  int       `json:"successful_generations"`
	FailedGenerations      int       `json:"failed_generations"`
	TotalCreditsUsed       int       `json:"total_credits_used"`
	TotalGenerationTime    int       `json:"total_generation_time"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type APIConfig struct {
	ID                   int       `json:"id"`
	UserID               int       `json:"user_id"`
	ConfigName           string    `json:"config_name"`
	APIKeyEncrypted      string    `json:"-"`
	APIBaseURL           string    `json:"api_base_url"`
	Model                string    `json:"model"`
	IsActive             bool      `json:"is_active"`
	MaxRequestsPerMinute int       `json:"max_requests_per_minute"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
