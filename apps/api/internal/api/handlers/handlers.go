// Package handlers는 HTTP 핸들러 구조체와 공통 헬퍼를 정의한다.
// 핸들러는 도메인 단위로 파일 분리: health, lab, user.
package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler는 모든 HTTP 핸들러가 공유하는 의존성을 묶은 구조체.
// New()에서 한 번 생성되어 router.go에서 각 엔드포인트에 메서드로 등록됨.
type Handler struct {
	log *zap.Logger
}

func New(log *zap.Logger) *Handler {
	return &Handler{log: log}
}

// errResp는 API 에러 응답의 공통 JSON 포맷.
// 프론트엔드 lib/api.ts의 ApiError 타입과 대응.
type errResp struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func (h *Handler) err(c *gin.Context, status int, msg string) {
	c.JSON(status, errResp{Error: msg})
}
