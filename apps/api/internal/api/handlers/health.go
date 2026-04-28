// 헬스체크 핸들러 — GET /health. 인증 없이 접근 가능.
// K8s readinessProbe/livenessProbe가 이 엔드포인트를 사용한다.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.1.0",
		"service": "cledyu-backend",
	})
}
