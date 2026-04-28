// Package middleware는 Gin 미들웨어 체인에 사용할 핸들러를 제공한다.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole은 JWT 미들웨어가 Context에 주입한 user_role을 검사한다.
// admin은 모든 역할 권한을 포함하므로 항상 통과.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get("user_role")
		if userRole != role && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}
