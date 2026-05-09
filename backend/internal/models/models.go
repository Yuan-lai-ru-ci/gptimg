package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	AvatarURL    string    `json:"avatar_url"`
	Credits      int       `json:"credits"`
	QuotaLimit   int       `json:"quota_limit"`
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
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	SessionID      string    `json:"session_id"`
	Prompt         string    `json:"prompt"`
	RevisedPrompt  string    `json:"revised_prompt"`
	ImageURL       string    `json:"image_url"`
	LocalPath      string    `json:"local_path"`
	Size           string    `json:"size"`
	Quality        string    `json:"quality"`
	Style          string    `json:"style"`
	Model          string    `json:"model"`
	CreditsUsed    int       `json:"credits_used"`
	GenerationTime int       `json:"generation_time"`
	Status         string    `json:"status"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}

type UsageStats struct {
	ID                    int       `json:"id"`
	UserID                int       `json:"user_id"`
	Date                  string    `json:"date"`
	TotalGenerations      int       `json:"total_generations"`
	SuccessfulGenerations int       `json:"successful_generations"`
	FailedGenerations     int       `json:"failed_generations"`
	TotalCreditsUsed      int       `json:"total_credits_used"`
	TotalGenerationTime   int       `json:"total_generation_time"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
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

type LLMConfig struct {
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

type AdminUserSummary struct {
	ID                    int       `json:"id"`
	Username              string    `json:"username"`
	Email                 string    `json:"email"`
	Credits               int       `json:"credits"`
	QuotaLimit            int       `json:"quota_limit"`
	Role                  string    `json:"role"`
	Status                string    `json:"status"`
	TotalGenerations      int       `json:"total_generations"`
	SuccessfulGenerations int       `json:"successful_generations"`
	TotalCreditsUsed      int       `json:"total_credits_used"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type AdminOverview struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	SuspendedUsers    int `json:"suspended_users"`
	TotalGenerations  int `json:"total_generations"`
	SuccessfulRuns    int `json:"successful_runs"`
	FailedRuns        int `json:"failed_runs"`
	TotalCreditsUsed  int `json:"total_credits_used"`
	ActiveAPIConfigs  int `json:"active_api_configs"`
	HealthyAPIConfigs int `json:"healthy_api_configs"`
}

type APIConfigStatus struct {
	ID                   int       `json:"id"`
	ConfigName           string    `json:"config_name"`
	APIBaseURL           string    `json:"api_base_url"`
	Model                string    `json:"model"`
	IsActive             bool      `json:"is_active"`
	MaxRequestsPerMinute int       `json:"max_requests_per_minute"`
	Healthy              bool      `json:"healthy"`
	StatusCode           int       `json:"status_code"`
	StatusMessage        string    `json:"status_message"`
	CheckedAt            time.Time `json:"checked_at"`
}

type LLMConfigStatus struct {
	ID                   int       `json:"id"`
	ConfigName           string    `json:"config_name"`
	APIBaseURL           string    `json:"api_base_url"`
	Model                string    `json:"model"`
	IsActive             bool      `json:"is_active"`
	MaxRequestsPerMinute int       `json:"max_requests_per_minute"`
	Healthy              bool      `json:"healthy"`
	StatusCode           int       `json:"status_code"`
	StatusMessage        string    `json:"status_message"`
	CheckedAt            time.Time `json:"checked_at"`
}

type PPTSlidePlan struct {
	SlideNumber     int    `json:"slide_number"`
	Title           string `json:"title"`
	Objective       string `json:"objective"`
	LayoutNotes     string `json:"layout_notes"`
	PageDescription string `json:"page_description"`
	ImagePrompt     string `json:"image_prompt"`
	SpeakerNotes    string `json:"speaker_notes"`
}

type PPTPlan struct {
	DeckTitle              string         `json:"deck_title"`
	DeckGoal               string         `json:"deck_goal"`
	VisualDirection        string         `json:"visual_direction"`
	MasterStyleDescription string         `json:"master_style_description"`
	ConsistencyRules       []string       `json:"consistency_rules"`
	GenerationMode         string         `json:"generation_mode,omitempty"`
	Slides                 []PPTSlidePlan `json:"slides"`
}
