package repository

import (
	"database/sql"
	"gptimg/internal/database"
	"gptimg/internal/models"
	"time"
)

type StatsRepository struct{}

func NewStatsRepository() *StatsRepository {
	return &StatsRepository{}
}

func (r *StatsRepository) UpsertDailyStats(userID int, date string, creditsUsed int, generationTime int, success bool) error {
	query := `INSERT INTO usage_stats (user_id, date, total_generations, successful_generations,
			  failed_generations, total_credits_used, total_generation_time)
			  VALUES (?, ?, 1, ?, ?, ?, ?)
			  ON CONFLICT(user_id, date) DO UPDATE SET
			  total_generations = total_generations + 1,
			  successful_generations = successful_generations + ?,
			  failed_generations = failed_generations + ?,
			  total_credits_used = total_credits_used + ?,
			  total_generation_time = total_generation_time + ?,
			  updated_at = ?`

	successCount := 0
	failCount := 0
	if success {
		successCount = 1
	} else {
		failCount = 1
	}

	_, err := database.DB.Exec(query, userID, date, successCount, failCount, creditsUsed, generationTime,
		successCount, failCount, creditsUsed, generationTime, time.Now())
	return err
}

func (r *StatsRepository) GetDailyStats(userID int, date string) (*models.UsageStats, error) {
	stats := &models.UsageStats{}
	query := `SELECT id, user_id, date, total_generations, successful_generations, failed_generations,
			  total_credits_used, total_generation_time, created_at, updated_at
			  FROM usage_stats WHERE user_id = ? AND date = ?`
	err := database.DB.QueryRow(query, userID, date).Scan(
		&stats.ID, &stats.UserID, &stats.Date, &stats.TotalGenerations,
		&stats.SuccessfulGenerations, &stats.FailedGenerations, &stats.TotalCreditsUsed,
		&stats.TotalGenerationTime, &stats.CreatedAt, &stats.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return stats, err
}

func (r *StatsRepository) GetStatsRange(userID int, startDate, endDate string) ([]*models.UsageStats, error) {
	query := `SELECT id, user_id, date, total_generations, successful_generations, failed_generations,
			  total_credits_used, total_generation_time, created_at, updated_at
			  FROM usage_stats WHERE user_id = ? AND date BETWEEN ? AND ? ORDER BY date DESC`
	rows, err := database.DB.Query(query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statsList []*models.UsageStats
	for rows.Next() {
		stats := &models.UsageStats{}
		err := rows.Scan(
			&stats.ID, &stats.UserID, &stats.Date, &stats.TotalGenerations,
			&stats.SuccessfulGenerations, &stats.FailedGenerations, &stats.TotalCreditsUsed,
			&stats.TotalGenerationTime, &stats.CreatedAt, &stats.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, stats)
	}
	return statsList, nil
}
