package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/store"
)

// URLUtil is used to help in- and outgoing requests for json
// un- and marshalling
type URLUtil struct {
	URL string `binding:"required"`
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
	user := c.MustGet("user").(*jwtClaims)
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
		return // return normal 404 error if such an error occurs
	} else if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if h.store.IncreaseVisitCounter(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, entry.Public.URL)
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(c *gin.Context) {
	var data URLUtil
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := c.MustGet("user").(*jwtClaims)
	id, err := h.store.CreateEntry(store.Entry{
		Public: store.EntryPublicData{
			URL: data.URL,
		},
		RemoteAddr:    c.ClientIP(),
		OAuthProvider: user.OAuthProvider,
		OAuthID:       user.OAuthID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data.URL = h.getSchemaAndHost(c) + "/" + id
	c.JSON(http.StatusOK, data)
}
