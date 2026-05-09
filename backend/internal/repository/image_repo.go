package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
)

type ImageRepository struct{}

func NewImageRepository() *ImageRepository {
	return &ImageRepository{}
}

func (r *ImageRepository) Create(record *models.ImageRecord) error {
	query := `INSERT INTO image_records (user_id, session_id, prompt, revised_prompt, image_url,
			  local_path, size, quality, style, model, credits_used, generation_time, status, error_message)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(query,
		record.UserID, record.SessionID, record.Prompt, record.RevisedPrompt, record.ImageURL,
		record.LocalPath, record.Size, record.Quality, record.Style, record.Model,
		record.CreditsUsed, record.GenerationTime, record.Status, record.ErrorMessage,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = int(id)
	return nil
}

func (r *ImageRepository) FindByID(id int) (*models.ImageRecord, error) {
	record := &models.ImageRecord{}
	query := `SELECT id, user_id, COALESCE(session_id,''), prompt, COALESCE(revised_prompt,''),
			  COALESCE(image_url,''), COALESCE(local_path,''), size, quality, COALESCE(style,''),
			  model, credits_used, COALESCE(generation_time,0), status, COALESCE(error_message,''), created_at
			  FROM image_records WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(
		&record.ID, &record.UserID, &record.SessionID, &record.Prompt, &record.RevisedPrompt,
		&record.ImageURL, &record.LocalPath, &record.Size, &record.Quality, &record.Style,
		&record.Model, &record.CreditsUsed, &record.GenerationTime, &record.Status,
		&record.ErrorMessage, &record.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return record, err
}

func (r *ImageRepository) FindBySessionID(sessionID string, limit, offset int) ([]*models.ImageRecord, error) {
	query := `SELECT id, user_id, COALESCE(session_id,''), prompt, COALESCE(revised_prompt,''),
			  COALESCE(image_url,''), COALESCE(local_path,''), size, quality, COALESCE(style,''),
			  model, credits_used, COALESCE(generation_time,0), status, COALESCE(error_message,''), created_at
			  FROM image_records WHERE session_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?`
	rows, err := database.DB.Query(query, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ImageRecord
	for rows.Next() {
		record := &models.ImageRecord{}
		err := rows.Scan(
			&record.ID, &record.UserID, &record.SessionID, &record.Prompt, &record.RevisedPrompt,
			&record.ImageURL, &record.LocalPath, &record.Size, &record.Quality, &record.Style,
			&record.Model, &record.CreditsUsed, &record.GenerationTime, &record.Status,
			&record.ErrorMessage, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *ImageRepository) FindByUserID(userID int, limit, offset int) ([]*models.ImageRecord, error) {
	query := `SELECT id, user_id, COALESCE(session_id,''), prompt, COALESCE(revised_prompt,''),
			  COALESCE(image_url,''), COALESCE(local_path,''), size, quality, COALESCE(style,''),
			  model, credits_used, COALESCE(generation_time,0), status, COALESCE(error_message,''), created_at
			  FROM image_records WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := database.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.ImageRecord
	for rows.Next() {
		record := &models.ImageRecord{}
		err := rows.Scan(
			&record.ID, &record.UserID, &record.SessionID, &record.Prompt, &record.RevisedPrompt,
			&record.ImageURL, &record.LocalPath, &record.Size, &record.Quality, &record.Style,
			&record.Model, &record.CreditsUsed, &record.GenerationTime, &record.Status,
			&record.ErrorMessage, &record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (r *ImageRepository) Delete(id int, userID int) error {
	query := `DELETE FROM image_records WHERE id = ? AND user_id = ?`
	_, err := database.DB.Exec(query, id, userID)
	return err
}
