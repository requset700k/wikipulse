// JWT 인증 미들웨어 — Authorization 헤더, Cookie, Query param 순서로 토큰을 추출.
// Query param은 WebSocket 연결 시 브라우저가 헤더를 못 보내기 때문에 필요.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kt-techuplabs/cledyu/backend/internal/config"
)

// JWT validates the access token from Authorization header or cookie.
// TODO: Week 3 — replace stub with real Keycloak JWKS validation.
func JWT(_ *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		// Stub: accept any non-empty token, inject mock claims.
		// admin role → instructor 엔드포인트 포함 전체 접근 가능 (dev only)
		c.Set("user_id", "mock-user-id")
		c.Set("user_email", "admin@kt.com")
		c.Set("user_name", "Dev Admin")
		c.Set("user_role", "admin")
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// 1. Authorization header (일반 API 요청)
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	// 2. Cookie (브라우저 — Keycloak 로그인 후)
	if cookie, err := c.Cookie("access_token"); err == nil {
		return cookie
	}
	// 3. Query param (WebSocket — 브라우저는 WS 헤더를 못 보냄)
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}
