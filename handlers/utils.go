package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getSchemaAndHost(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}
