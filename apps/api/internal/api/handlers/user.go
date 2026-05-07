package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMe는 JWT 미들웨어가 Context에 주입한 사용자 정보를 반환한다.
// 현재는 mock claims 반환. Phase D(Keycloak 연동) 완료 후 실 DB 조회로 교체.
func (h *Handler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     c.GetString("user_id"),
		"email":  c.GetString("user_email"),
		"name":   c.GetString("user_name"),
		"role":   c.GetString("user_role"),
		"points": 0,
		"badges": []gin.H{},
	})
}
