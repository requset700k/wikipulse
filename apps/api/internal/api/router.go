// Package api는 Gin HTTP 라우터와 미들웨어 체인을 구성한다.
package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kt-techuplabs/cledyu/backend/internal/api/handlers"
	"github.com/kt-techuplabs/cledyu/backend/internal/api/middleware"
	"github.com/kt-techuplabs/cledyu/backend/internal/config"
	"github.com/kt-techuplabs/cledyu/backend/internal/service"
	"go.uber.org/zap"
)

// NewRouter는 라우터를 생성한다.
// 미들웨어 적용 순서: Recovery → Logger → CORS → (JWT → RequireRole 순으로 protected 그룹에만)
func NewRouter(cfg *config.Config, log *zap.Logger, sessions *service.SessionService, hints *service.HintService) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()
	r.Use(gin.Recovery())        // 패닉 발생 시 500 응답 후 서버 유지
	r.Use(middleware.Logger(log)) // 모든 요청/응답 구조화 로그
	r.Use(cors.New(cors.Config{
		// 로컬 개발 시 Next.js dev server(3000)에서의 요청 허용.
		// 프로덕션에서는 Kong Gateway가 CORS를 처리하므로 이 설정은 dev 전용.
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	h := handlers.New(cfg, log, sessions, hints)

	// 인증 불필요 — 헬스체크와 API 문서
	r.GET("/health", h.Health)
	r.GET("/api/docs", h.SwaggerUI)
	r.GET("/api/docs/openapi.yaml", h.OpenAPISpec)

	v1 := r.Group("/api/v1")

	// 인증 불필요 — Keycloak OIDC 흐름
	authGroup := v1.Group("/auth")
	{
		authGroup.GET("/login", h.Login)       // Keycloak 인증 페이지로 리다이렉트
		authGroup.GET("/callback", h.Callback) // 인증 코드 → JWT 교환 (Week 3)
		authGroup.POST("/logout", h.Logout)    // access_token 쿠키 삭제
	}

	// JWT 미들웨어 적용 — 이 그룹 이하는 모두 토큰 필요
	protected := v1.Group("")
	protected.Use(middleware.JWT(cfg))
	{
		protected.GET("/me", h.GetMe)
		protected.GET("/leaderboard", h.GetLeaderboard)
		protected.GET("/me/badges", h.GetMyBadges)

		labs := protected.Group("/labs")
		{
			labs.GET("", h.ListLabs)
			labs.GET("/:id", h.GetLab)
		}

		sessions := protected.Group("/sessions")
		{
			sessions.POST("", h.CreateSession)            // VM 프로비저닝 시작
			sessions.GET("/:id", h.GetSession)            // 세션 상태 폴링용
			sessions.DELETE("/:id", h.DeleteSession)      // 세션 종료 + VM destroy
			sessions.GET("/:id/ws", h.TerminalWS)         // xterm.js WebSocket 연결점
			sessions.POST("/:id/validate", h.TriggerValidation) // 단계 완료 검증 요청
			sessions.POST("/:id/hint", h.RequestHint)     // AI 힌트 요청 (Rate Limit 적용)
			sessions.GET("/:id/steps", h.GetSteps)
			sessions.PUT("/:id/steps/:stepId", h.UpdateStep)
		}

		// RequireRole: instructor 또는 admin만 접근 가능
		instructor := protected.Group("/instructor")
		instructor.Use(middleware.RequireRole("instructor"))
		{
			instructor.GET("/sessions", h.InstructorListSessions)
			instructor.GET("/sessions/:id/ws", h.InstructorTerminalWS) // 수강생 터미널 관전 (읽기 전용)
			instructor.POST("/sessions/:id/inject", h.InstructorInjectCommand) // 수강생 터미널에 명령 주입
		}
	}

	return r
}
