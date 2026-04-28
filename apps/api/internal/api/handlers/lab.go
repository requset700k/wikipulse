// Lab 목록/상세 핸들러. 현재는 labstore(메모리 mock)에서 조회.
// 추후 DB + Lab DSL YAML 기반 저장소로 교체 예정 (김찬영 담당).
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kt-techuplabs/cledyu/backend/internal/api/handlers/internal/labstore"
)

func (h *Handler) ListLabs(c *gin.Context) {
	labs := labstore.All()
	c.JSON(http.StatusOK, gin.H{
		"items": labs,
		"total": len(labs),
	})
}

func (h *Handler) GetLab(c *gin.Context) {
	lab, ok := labstore.Find(c.Param("id"))
	if !ok {
		h.err(c, http.StatusNotFound, "lab not found")
		return
	}
	c.JSON(http.StatusOK, lab)
}
