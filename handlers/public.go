package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/handlers/auth"
	"github.com/maxibanki/golang-url-shortener/store"
)

// urlUtil is used to help in- and outgoing requests for json
// un- and marshalling
type urlUtil struct {
	URL        string `binding:"required"`
	ID         string
	Expiration time.Time
}

// handleLookup is the http handler for getting the infos
func (h *Handler) handleLookup(c *gin.Context) {
	var data struct {
		ID string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.store.GetEntryByID(data.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	user := c.MustGet("user").(*auth.JWTClaims)
	if entry.OAuthID != user.OAuthID || entry.OAuthProvider != user.OAuthProvider {
		c.JSON(http.StatusOK, store.Entry{
			Public: store.EntryPublicData{
				URL: entry.Public.URL,
			},
		})
		return
	}
	c.JSON(http.StatusOK, entry.Public)
}

// handleAccess handles the access for incoming requests
func (h *Handler) handleAccess(c *gin.Context) {
	var id string
	if len(c.Request.URL.Path) > 1 {
		id = c.Request.URL.Path[1:]
	}
	entry, err := h.store.GetEntryByID(id)
	if err == store.ErrIDIsEmpty || err == store.ErrNoEntryFound {
		return
	} else if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if time.Now().After(entry.Public.Expiration) && !entry.Public.Expiration.IsZero() {
		return
	}
	if err := h.store.IncreaseVisitCounter(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, entry.Public.URL)
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(c *gin.Context) {
	var data urlUtil
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := c.MustGet("user").(*auth.JWTClaims)
	id, err := h.store.CreateEntry(store.Entry{
		Public: store.EntryPublicData{
			URL:        data.URL,
			Expiration: data.Expiration,
		},
		RemoteAddr:    c.ClientIP(),
		OAuthProvider: user.OAuthProvider,
		OAuthID:       user.OAuthID,
	}, data.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data.URL = h.getSchemaAndHost(c) + "/" + id
	c.JSON(http.StatusOK, data)
}

func (h *Handler) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"providers": h.providers})
}

func (h *Handler) getSchemaAndHost(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}
