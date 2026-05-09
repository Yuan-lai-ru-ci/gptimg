package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
	"time"
)

type SessionRepository struct{}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

func (r *SessionRepository) Create(session *models.Session) error {
	query := `INSERT INTO sessions (id, user_id, title, last_message_at, message_count)
			  VALUES (?, ?, ?, ?, ?)`
	_, err := database.DB.Exec(query, session.ID, session.UserID, session.Title,
		session.LastMessageAt, session.MessageCount)
	return err
}

func (r *SessionRepository) FindByID(id string) (*models.Session, error) {
	session := &models.Session{}
	var lastMsg, createdAt, updatedAt string
	query := `SELECT id, user_id, COALESCE(title,''), COALESCE(last_message_at,''), message_count,
			  COALESCE(created_at,''), COALESCE(updated_at,'')
			  FROM sessions WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(
		&session.ID, &session.UserID, &session.Title, &lastMsg,
		&session.MessageCount, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	session.LastMessageAt = parseTime(lastMsg)
	session.CreatedAt = parseTime(createdAt)
	session.UpdatedAt = parseTime(updatedAt)
	return session, nil
}

func (r *SessionRepository) FindByUserID(userID int, limit, offset int) ([]*models.Session, error) {
	query := `SELECT id, user_id, COALESCE(title,''), COALESCE(last_message_at,''), message_count,
			  COALESCE(created_at,''), COALESCE(updated_at,'')
			  FROM sessions WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	rows, err := database.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		var lastMsg, createdAt, updatedAt string
		err := rows.Scan(
			&session.ID, &session.UserID, &session.Title, &lastMsg,
			&session.MessageCount, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}
		session.LastMessageAt = parseTime(lastMsg)
		session.CreatedAt = parseTime(createdAt)
		session.UpdatedAt = parseTime(updatedAt)
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (r *SessionRepository) Update(session *models.Session) error {
	query := `UPDATE sessions SET title = ?, last_message_at = ?, message_count = ?, updated_at = ?
			  WHERE id = ?`
	_, err := database.DB.Exec(query, session.Title, session.LastMessageAt,
		session.MessageCount, time.Now(), session.ID)
	return err
}

func (r *SessionRepository) Delete(id string, userID int) error {
	query := `DELETE FROM sessions WHERE id = ? AND user_id = ?`
	_, err := database.DB.Exec(query, id, userID)
	return err
}
