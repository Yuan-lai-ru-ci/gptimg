package handlers

import (
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsRepo *repository.StatsRepository
}

func NewStatsHandler(statsRepo *repository.StatsRepository) *StatsHandler {
	return &StatsHandler{
		statsRepo: statsRepo,
	}
}

func (h *StatsHandler) GetOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")

	today := time.Now().Format("2006-01-02")
	stats, err := h.statsRepo.GetDailyStats(userID.(int), today)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	if stats == nil {
		stats = &models.UsageStats{
			UserID:                 userID.(int),
			Date:                   today,
			TotalGenerations:       0,
			SuccessfulGenerations:  0,
			FailedGenerations:      0,
			TotalCreditsUsed:       0,
			TotalGenerationTime:    0,
		}
	}

	response.Success(c, stats)
}

func (h *StatsHandler) GetDaily(c *gin.Context) {
	userID, _ := c.Get("user_id")

	days := c.DefaultQuery("days", "7")
	daysInt := 7
	if d, err := strconv.Atoi(days); err == nil && d > 0 && d <= 90 {
		daysInt = d
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -daysInt+1)

	statsList, err := h.statsRepo.GetStatsRange(
		userID.(int),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, statsList)
}
