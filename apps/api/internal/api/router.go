// Package api wires the Gin router and middleware chain.
// 미들웨어 적용 순서: Recovery → Logger → CORS → JWT (protected 그룹만)
package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/requset700k/cledyu/api/internal/api/handlers"
	"github.com/requset700k/cledyu/api/internal/config"
	"github.com/requset700k/cledyu/api/internal/middleware"
	"go.uber.org/zap"
)

func NewRouter(cfg *config.Config, log *zap.Logger) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()
	r.Use(gin.Recovery())         // 패닉 발생 시 500 반환 후 서버 유지
	r.Use(middleware.Logger(log)) // 모든 요청/응답 구조화 로그
	r.Use(cors.New(cors.Config{
		// Next.js dev server(3000) 및 클러스터 프론트엔드에서의 요청 허용.
		// 프로덕션에서는 Traefik이 CORS를 처리하므로 이 설정은 로컬 개발 전용.
		AllowOrigins:     []string{"http://localhost:3000", "https://app.cledyu.local"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	h := handlers.New(log)

	// 인증 불필요 — 헬스체크
	r.GET("/health", h.Health)

	// JWT 미들웨어 적용 — 이 그룹 이하는 모두 토큰 필요.
	// TODO: Phase D(Keycloak 연동) 완료 후 stub → 실 JWKS 검증으로 교체.
	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWT())
	{
		v1.GET("/me", h.GetMe)          // 현재 로그인 사용자 정보 (mock)
		v1.GET("/labs", h.ListLabs)     // Lab 목록 (mock)
		v1.GET("/labs/:id", h.GetLab)   // Lab 단건 조회 (mock)
	}

	return r
}
