// Package handlers는 HTTP 핸들러 구조체와 공통 헬퍼를 정의한다.
// 핸들러는 파일별로 도메인 단위로 분리: auth, session, terminal, lab, instructor, docs, gamification
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kt-techuplabs/cledyu/backend/internal/config"
	termreg "github.com/kt-techuplabs/cledyu/backend/internal/infra/terminal"
	"github.com/kt-techuplabs/cledyu/backend/internal/service"
	"go.uber.org/zap"
)

// Handler는 모든 HTTP 핸들러가 공유하는 의존성을 묶은 구조체.
// New()에서 한 번 생성되어 router.go에서 각 엔드포인트에 메서드로 등록됨.
type Handler struct {
	cfg       *config.Config
	log       *zap.Logger
	sessions  *service.SessionService
	hints     *service.HintService
	terminals *termreg.Registry // 활성 PTY 세션 레지스트리 (강사 관전에 사용)
}

// New는 Handler를 생성한다. terminals는 프로세스 전역 Global 레지스트리를 사용.
func New(cfg *config.Config, log *zap.Logger, sessions *service.SessionService, hints *service.HintService) *Handler {
	return &Handler{
		cfg:       cfg,
		log:       log,
		sessions:  sessions,
		hints:     hints,
		terminals: termreg.Global,
	}
}

// errResp는 API 에러 응답의 공통 JSON 포맷. 프론트의 ApiError 타입과 대응.
type errResp struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func (h *Handler) err(c *gin.Context, status int, msg string) {
	c.JSON(status, errResp{Error: msg})
}

// notImplemented는 Week 3 이후 구현 예정인 엔드포인트에 사용.
func (h *Handler) notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, errResp{
		Error: "not implemented",
		Code:  "NOT_IMPLEMENTED",
	})
}
