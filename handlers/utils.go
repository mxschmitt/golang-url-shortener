package handlers

import (
	"crypto/rand"
	"encoding/base64"
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

func (h *Handler) randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
