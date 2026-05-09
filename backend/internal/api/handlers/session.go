package handlers

import (
	"crypto/rand"
	"fmt"
	"gptimg/internal/models"
	"gptimg/internal/repository"
	"gptimg/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type SessionHandler struct {
	sessionRepo *repository.SessionRepository
	imageRepo   *repository.ImageRepository
}

func NewSessionHandler(sessionRepo *repository.SessionRepository, imageRepo *repository.ImageRepository) *SessionHandler {
	return &SessionHandler{
		sessionRepo: sessionRepo,
		imageRepo:   imageRepo,
	}
}

type CreateSessionRequest struct {
	Title string `json:"title"`
}

type UpdateSessionRequest struct {
	Title string `json:"title"`
}

func (h *SessionHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	session := &models.Session{
		ID:            generateID(),
		UserID:        userID.(int),
		Title:         req.Title,
		LastMessageAt: time.Now(),
		MessageCount:  0,
	}

	if err := h.sessionRepo.Create(session); err != nil {
		response.InternalError(c, "Failed to create session")
		return
	}

	response.Success(c, session)
}

func (h *SessionHandler) GetList(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	sessions, err := h.sessionRepo.FindByUserID(userID.(int), limit, offset)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, sessions)
}

func (h *SessionHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID := c.Param("id")

	session, err := h.sessionRepo.FindByID(sessionID)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if session == nil {
		response.NotFound(c, "Session not found")
		return
	}

	if session.UserID != userID.(int) {
		response.Forbidden(c, "Access denied")
		return
	}

	response.Success(c, session)
}

func (h *SessionHandler) GetMessages(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID := c.Param("id")

	session, err := h.sessionRepo.FindByID(sessionID)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if session == nil {
		response.NotFound(c, "Session not found")
		return
	}

	if session.UserID != userID.(int) {
		response.Forbidden(c, "Access denied")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.imageRepo.FindBySessionID(sessionID, limit, offset)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}

	response.Success(c, messages)
}

func (h *SessionHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID := c.Param("id")

	if err := h.sessionRepo.Delete(sessionID, userID.(int)); err != nil {
		response.InternalError(c, "Failed to delete session")
		return
	}

	response.Success(c, gin.H{"message": "Session deleted successfully"})
}

func (h *SessionHandler) Update(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID := c.Param("id")

	session, err := h.sessionRepo.FindByID(sessionID)
	if err != nil {
		response.InternalError(c, "Database error")
		return
	}
	if session == nil || session.UserID != userID.(int) {
		response.Forbidden(c, "Access denied")
		return
	}

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	session.Title = req.Title
	if err := h.sessionRepo.Update(session); err != nil {
		response.InternalError(c, "Failed to update session")
		return
	}

	response.Success(c, session)
}
