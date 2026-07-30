package handler

import (
	"net/http"
	"os"

	"go-gin-layered-architecture/internal/config"

	"github.com/gin-gonic/gin"
)

type DocsHandler struct {
	docsConfig *config.DocsConfig
}

func NewDocsHandler(docsConfig *config.DocsConfig) *DocsHandler {
	return &DocsHandler{docsConfig: docsConfig}
}

func (h *DocsHandler) Auth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		h.docsConfig.User: h.docsConfig.Password,
	})
}

func (h *DocsHandler) GetDocs(c *gin.Context) {
	spec, err := os.ReadFile("docs/swagger.json")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	scalarJS, err := os.ReadFile("docs/static/scalar.js")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!doctype html>
		<html>
		<head><title>API Docs</title><meta charset="utf-8"></head>
		<body>
		<script id="api-reference" type="application/json">`+string(spec)+`</script>
		<script>`+string(scalarJS)+`</script>
		</body>
		</html>`))
}
