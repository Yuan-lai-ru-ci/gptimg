package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
	"time"
)

type LLMConfigRepository struct{}

func NewLLMConfigRepository() *LLMConfigRepository {
	return &LLMConfigRepository{}
}

func (r *LLMConfigRepository) Create(config *models.LLMConfig) error {
	query := `INSERT INTO llm_configs (user_id, config_name, api_key_encrypted, api_base_url, model, is_active, max_requests_per_minute)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(query, config.UserID, config.ConfigName, config.APIKeyEncrypted, config.APIBaseURL, config.Model, config.IsActive, config.MaxRequestsPerMinute)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	config.ID = int(id)
	return nil
}

func (r *LLMConfigRepository) FindByID(id int) (*models.LLMConfig, error) {
	config := &models.LLMConfig{}
	query := `SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model, is_active, max_requests_per_minute, created_at, updated_at
			  FROM llm_configs WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(
		&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
		&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return config, err
}

func (r *LLMConfigRepository) FindAll() ([]*models.LLMConfig, error) {
	rows, err := database.DB.Query(`SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model, is_active, max_requests_per_minute, created_at, updated_at FROM llm_configs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.LLMConfig
	for rows.Next() {
		config := &models.LLMConfig{}
		if err := rows.Scan(
			&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
			&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *LLMConfigRepository) FindActiveConfigs() ([]*models.LLMConfig, error) {
	rows, err := database.DB.Query(`SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model, is_active, max_requests_per_minute, created_at, updated_at FROM llm_configs WHERE is_active = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.LLMConfig
	for rows.Next() {
		config := &models.LLMConfig{}
		if err := rows.Scan(
			&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
			&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
			&config.CreatedAt, &config.UpdatedAt,
		); err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *LLMConfigRepository) Update(config *models.LLMConfig) error {
	query := `UPDATE llm_configs SET config_name = ?, api_key_encrypted = ?, api_base_url = ?, model = ?, is_active = ?, max_requests_per_minute = ?, updated_at = ? WHERE id = ?`
	_, err := database.DB.Exec(query, config.ConfigName, config.APIKeyEncrypted, config.APIBaseURL, config.Model, config.IsActive, config.MaxRequestsPerMinute, time.Now(), config.ID)
	return err
}

func (r *LLMConfigRepository) Delete(id int) error {
	_, err := database.DB.Exec(`DELETE FROM llm_configs WHERE id = ?`, id)
	return err
}
