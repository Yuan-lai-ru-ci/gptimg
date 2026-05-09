package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
	"time"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (username, email, password_hash, credits, quota_limit, role, status)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(query, user.Username, user.Email, user.PasswordHash,
		user.Credits, user.QuotaLimit, user.Role, user.Status)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, COALESCE(avatar_url,''), credits, COALESCE(quota_limit,50), role, status,
			  created_at, updated_at FROM users WHERE email = ?`
	err := database.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Credits, &user.QuotaLimit, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, COALESCE(avatar_url,''), credits, COALESCE(quota_limit,50), role, status,
			  created_at, updated_at FROM users WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Credits, &user.QuotaLimit, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) UpdateCredits(userID int, credits int) error {
	query := `UPDATE users SET credits = ?, updated_at = ? WHERE id = ?`
	_, err := database.DB.Exec(query, credits, time.Now(), userID)
	return err
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, COALESCE(avatar_url,''), credits, COALESCE(quota_limit,50), role, status,
			  created_at, updated_at FROM users WHERE username = ?`
	err := database.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Credits, &user.QuotaLimit, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) ListWithUsage() ([]*models.AdminUserSummary, error) {
	query := `SELECT
		u.id,
		u.username,
		u.email,
		u.credits,
		COALESCE(u.quota_limit, 50),
		u.role,
		u.status,
		COALESCE(SUM(s.total_generations), 0),
		COALESCE(SUM(s.successful_generations), 0),
		COALESCE(SUM(s.total_credits_used), 0),
		u.created_at,
		u.updated_at
	FROM users u
	LEFT JOIN usage_stats s ON s.user_id = u.id
	GROUP BY u.id
	ORDER BY CASE WHEN u.role = 'admin' THEN 0 ELSE 1 END, u.created_at DESC`
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.AdminUserSummary
	for rows.Next() {
		user := &models.AdminUserSummary{}
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Credits,
			&user.QuotaLimit,
			&user.Role,
			&user.Status,
			&user.TotalGenerations,
			&user.SuccessfulGenerations,
			&user.TotalCreditsUsed,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) UpdateAdminFields(userID, credits, quotaLimit int, role, status string) error {
	query := `UPDATE users SET credits = ?, quota_limit = ?, role = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := database.DB.Exec(query, credits, quotaLimit, role, status, time.Now(), userID)
	return err
}

func (r *UserRepository) DeleteUserWithData(userID int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := []string{
		`DELETE FROM image_records WHERE user_id = ?`,
		`DELETE FROM usage_stats WHERE user_id = ?`,
		`DELETE FROM sessions WHERE user_id = ?`,
		`DELETE FROM users WHERE id = ?`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = ?, email = ?, avatar_url = ?, updated_at = ? WHERE id = ?`
	_, err := database.DB.Exec(query, user.Username, user.Email, user.AvatarURL, time.Now(), user.ID)
	return err
}
