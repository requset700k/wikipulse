package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health는 서버 상태를 반환한다.
// ArgoCD readinessProbe 및 모니터링 헬스체크 용도.
// TODO: Phase B(Redis 연동) 완료 후 Redis PING 결과를 응답에 포함.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.1.0",
	})
}
