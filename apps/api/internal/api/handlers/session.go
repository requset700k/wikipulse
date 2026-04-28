// Lab 세션 관련 핸들러 — 생성, 조회, 삭제, 단계 진행, 검증, AI 힌트.
// 세션 생성 시 VM 프로비저닝이 비동기로 시작되고, 프론트는 GET /sessions/:id를 폴링하여 status를 확인.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kt-techuplabs/cledyu/backend/internal/api/handlers/internal/labstore"
	"github.com/kt-techuplabs/cledyu/backend/internal/domain/session"
	"github.com/kt-techuplabs/cledyu/backend/internal/service"
	"go.uber.org/zap"
)

func (h *Handler) CreateSession(c *gin.Context) {
	var req struct {
		LabID string `json:"lab_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	lab, ok := labstore.Find(req.LabID)
	if !ok {
		h.err(c, http.StatusNotFound, "lab not found")
		return
	}

	sess, err := h.sessions.Create(c.Request.Context(), service.CreateParams{
		UserID:    c.GetString("user_id"),
		LabID:     req.LabID,
		VMType:    lab.VMType,
		StepCount: lab.StepCount,
	})
	if err != nil {
		h.log.Error("create session", zap.Error(err))
		h.err(c, http.StatusInternalServerError, "failed to create session")
		return
	}
	c.JSON(http.StatusCreated, sess)
}

func (h *Handler) GetSession(c *gin.Context) {
	sess, err := h.sessions.Get(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if err != nil {
		h.err(c, http.StatusNotFound, "session not found")
		return
	}
	c.JSON(http.StatusOK, sess)
}

func (h *Handler) DeleteSession(c *gin.Context) {
	if err := h.sessions.Delete(c.Request.Context(), c.Param("id"), c.GetString("user_id")); err != nil {
		h.err(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetSteps(c *gin.Context) {
	steps, err := h.sessions.GetSteps(c.Request.Context(), c.Param("id"), c.GetString("user_id"))
	if err != nil {
		h.err(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": steps, "total": len(steps)})
}

func (h *Handler) UpdateStep(c *gin.Context) {
	stepID, err := strconv.Atoi(c.Param("stepId"))
	if err != nil {
		h.err(c, http.StatusBadRequest, "invalid step id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.sessions.UpdateStep(
		c.Request.Context(), c.Param("id"), c.GetString("user_id"),
		stepID, session.StepStatus(req.Status),
	); err != nil {
		h.err(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"step_id": stepID, "status": req.Status})
}

func (h *Handler) TriggerValidation(c *gin.Context) {
	var req struct {
		StepID int `json:"step_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	sessionID := c.Param("id")

	// TODO: Week 7 — publish to Kafka validation-requests (김찬영 Validation Engine)
	// For now: simulate validation passing after 2s
	h.sessions.SimulateValidation(sessionID, req.StepID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "pending",
		"message": "검증 중... (2초 후 결과 확인)",
	})
}

func (h *Handler) RequestHint(c *gin.Context) {
	var req struct {
		StepID          int    `json:"step_id" binding:"required"`
		HintLevel       int    `json:"hint_level" binding:"required,min=1,max=3"`
		TerminalHistory string `json:"terminal_history"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.err(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.hints.Request(c.Request.Context(), service.HintParams{
		UserID:          c.GetString("user_id"),
		SessionID:       c.Param("id"),
		StepID:          req.StepID,
		HintLevel:       req.HintLevel,
		TerminalHistory: req.TerminalHistory,
	})
	if errors.Is(err, service.ErrRateLimitExceeded) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":           "rate limit exceeded",
			"hints_remaining": 0,
		})
		return
	}
	if err != nil {
		h.err(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}
