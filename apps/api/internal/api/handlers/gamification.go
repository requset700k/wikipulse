// Gamification 핸들러 — 리더보드, 배지 (Month 2 구현 예정).
// 현재는 mock 데이터 반환. PostgreSQL 연동 및 포인트 계산 로직은 추후 추가.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetLeaderboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{
			{"rank": 1, "name": "홍길동", "points": 1250, "badges": 8},
			{"rank": 2, "name": "김철수", "points": 980, "badges": 6},
			{"rank": 3, "name": "이영희", "points": 870, "badges": 5},
		},
		"total": 3,
	})
}

func (h *Handler) GetMyBadges(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"items": []gin.H{},
		"total": 0,
	})
}
