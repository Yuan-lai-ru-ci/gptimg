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
	query := `INSERT INTO users (username, email, password_hash, credits, role, status)
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(query, user.Username, user.Email, user.PasswordHash,
		user.Credits, user.Role, user.Status)
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
	query := `SELECT id, username, email, password_hash, COALESCE(avatar_url,''), credits, role, status,
			  created_at, updated_at FROM users WHERE email = ?`
	err := database.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Credits, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, COALESCE(avatar_url,''), credits, role, status,
			  created_at, updated_at FROM users WHERE id = ?`
	err := database.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.AvatarURL,
		&user.Credits, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
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

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = ?, email = ?, avatar_url = ?, updated_at = ? WHERE id = ?`
	_, err := database.DB.Exec(query, user.Username, user.Email, user.AvatarURL, time.Now(), user.ID)
	return err
}
