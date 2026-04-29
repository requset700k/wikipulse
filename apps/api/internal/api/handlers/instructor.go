// 강사 전용 핸들러 — instructor 또는 admin 역할만 접근 가능 (RequireRole 미들웨어).
// InstructorListSessions: DB의 활성 세션 목록 + PTY 레지스트리에서 터미널 활성화 여부를 합쳐 반환.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) InstructorListSessions(c *gin.Context) {
	list, err := h.sessions.ActiveSessions(c.Request.Context())
	if err != nil {
		h.log.Error("list sessions", zap.Error(err))
		h.err(c, http.StatusInternalServerError, "failed to list sessions")
		return
	}

	// PTY registry에서 실제 터미널이 활성화된 세션에 has_terminal 플래그 추가
	h.log.Info("instructor sessions check", zap.Int("count", len(list)))
	type enriched struct {
		ID          string      `json:"id"`
		UserName    string      `json:"user_name"`
		UserEmail   string      `json:"user_email"`
		LabID       string      `json:"lab_id"`
		Status      string      `json:"status"`
		CurrentStep int         `json:"current_step"`
		StepsDone   int         `json:"steps_done"`
		StartedAt   interface{} `json:"started_at"`
		HasTerminal bool        `json:"has_terminal"`
	}

	result := make([]enriched, 0, len(list))
	for _, s := range list {
		_, hasTerminal := h.terminals.Get(s.ID)
		if hasTerminal {
			h.log.Info("active terminal found", zap.String("session", s.ID))
		}
		result = append(result, enriched{
			ID:          s.ID,
			UserName:    s.UserName,
			UserEmail:   s.UserEmail,
			LabID:       s.LabID,
			Status:      s.Status,
			CurrentStep: s.CurrentStep,
			StepsDone:   s.StepsDone,
			StartedAt:   s.StartedAt,
			HasTerminal: hasTerminal,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": result, "total": len(result)})
}
