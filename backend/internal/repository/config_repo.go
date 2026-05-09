package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
	"time"
)

type ConfigRepository struct{}

func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

func (r *ConfigRepository) Create(config *models.APIConfig) error {
	query := `INSERT INTO api_configs (user_id, config_name, api_key_encrypted, api_base_url,
			  model, is_active, max_requests_per_minute)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(query, config.UserID, config.ConfigName, config.APIKeyEncrypted,
		config.APIBaseURL, config.Model, config.IsActive, config.MaxRequestsPerMinute)
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

func (r *ConfigRepository) FindByID(id int) (*models.APIConfig, error) {
	config := &models.APIConfig{}
	query := `SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model,
			  is_active, max_requests_per_minute, created_at, updated_at
			  FROM api_configs WHERE id = ?`
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

func (r *ConfigRepository) FindActive() (*models.APIConfig, error) {
	config := &models.APIConfig{}
	query := `SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model,
			  is_active, max_requests_per_minute, created_at, updated_at
			  FROM api_configs WHERE is_active = 1 LIMIT 1`
	err := database.DB.QueryRow(query).Scan(
		&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
		&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return config, err
}

func (r *ConfigRepository) FindActiveConfigs() ([]*models.APIConfig, error) {
	query := `SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model,
			  is_active, max_requests_per_minute, created_at, updated_at
			  FROM api_configs WHERE is_active = 1 ORDER BY created_at DESC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.APIConfig
	for rows.Next() {
		config := &models.APIConfig{}
		err := rows.Scan(
			&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
			&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *ConfigRepository) FindAll() ([]*models.APIConfig, error) {
	query := `SELECT id, user_id, config_name, api_key_encrypted, api_base_url, model,
			  is_active, max_requests_per_minute, created_at, updated_at
			  FROM api_configs ORDER BY created_at DESC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*models.APIConfig
	for rows.Next() {
		config := &models.APIConfig{}
		err := rows.Scan(
			&config.ID, &config.UserID, &config.ConfigName, &config.APIKeyEncrypted,
			&config.APIBaseURL, &config.Model, &config.IsActive, &config.MaxRequestsPerMinute,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (r *ConfigRepository) Update(config *models.APIConfig) error {
	query := `UPDATE api_configs SET config_name = ?, api_key_encrypted = ?, api_base_url = ?,
			  model = ?, is_active = ?, max_requests_per_minute = ?, updated_at = ?
			  WHERE id = ?`
	_, err := database.DB.Exec(query, config.ConfigName, config.APIKeyEncrypted, config.APIBaseURL,
		config.Model, config.IsActive, config.MaxRequestsPerMinute, time.Now(), config.ID)
	return err
}

func (r *ConfigRepository) Delete(id int) error {
	query := `DELETE FROM api_configs WHERE id = ?`
	_, err := database.DB.Exec(query, id)
	return err
}
