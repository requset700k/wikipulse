// API 문서 핸들러 — Swagger UI와 openapi.yaml 파일 제공.
// GET /api/docs → Swagger UI (브라우저에서 API 탐색/테스트 가능)
// GET /api/docs/openapi.yaml → OpenAPI 스펙 파일 직접 다운로드
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>KT Tech-Up Labs — API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; }</style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload = function() {
  SwaggerUIBundle({
    url: "/api/docs/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis],
    layout: "BaseLayout",
    tryItOutEnabled: true,
  });
};
</script>
</body>
</html>`

func (h *Handler) SwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
}

func (h *Handler) OpenAPISpec(c *gin.Context) {
	c.File("openapi.yaml")
}
