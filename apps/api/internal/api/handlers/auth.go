// Keycloak OIDC 인증 핸들러 — 로그인, 콜백, 로그아웃.
// Week 3에 Callback 구현 예정. 현재는 stub JWT 미들웨어로 인증 우회.
package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Login(c *gin.Context) {
	if h.cfg.Auth.KeycloakURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "auth not configured",
			"note":  "set keycloak_url in config.yaml (Week 3)",
		})
		return
	}
	authURL := fmt.Sprintf(
		"%s/realms/%s/protocol/openid-connect/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+profile+email",
		h.cfg.Auth.KeycloakURL,
		h.cfg.Auth.KeycloakRealm,
		h.cfg.Auth.ClientID,
		url.QueryEscape(h.cfg.Auth.RedirectURL),
	)
	c.Redirect(http.StatusFound, authURL)
}

func (h *Handler) Callback(c *gin.Context) {
	// TODO: Week 3 — exchange code for JWT, set HTTP-only cookie, redirect to /labs
	h.notImplemented(c)
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *Handler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id":     c.GetString("user_id"),
		"email":  c.GetString("user_email"),
		"name":   c.GetString("user_name"),
		"role":   c.GetString("user_role"),
		"points": 0,
	})
}
