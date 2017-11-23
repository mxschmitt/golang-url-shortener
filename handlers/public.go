package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/handlers/auth"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/maxibanki/golang-url-shortener/util"
)

// urlUtil is used to help in- and outgoing requests for json
// un- and marshalling
type urlUtil struct {
	URL             string `binding:"required"`
	ID, DeletionURL string
	Expiration      *time.Time `json:",omitempty"`
}

// handleLookup is the http handler for getting the infos
func (h *Handler) handleLookup(c *gin.Context) {
	var data struct {
		ID string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.store.GetEntryByID(data.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if !h.oAuthPropertiesEquals(c, entry.OAuthID, entry.OAuthProvider) {
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
	url, err := h.store.GetURLAndIncrease(c.Request.URL.Path[1:])
	if err == store.ErrNoEntryFound {
		return
	} else if err != nil {
		http.Error(c.Writer, fmt.Sprintf("could not get and crease visitor counter: %v, ", err), http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(c *gin.Context) {
	var data urlUtil
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := c.MustGet("user").(*auth.JWTClaims)
	id, delID, err := h.store.CreateEntry(store.Entry{
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
	originURL := h.getURLOrigin(c)
	c.JSON(http.StatusOK, urlUtil{
		URL:         fmt.Sprintf("%s/%s", originURL, id),
		DeletionURL: fmt.Sprintf("%s/d/%s/%s", originURL, id, url.QueryEscape(base64.RawURLEncoding.EncodeToString(delID))),
	})
}

func (h *Handler) handleInfo(c *gin.Context) {
	info := gin.H{
		"providers": h.providers,
		"go":        runtime.Version(),
	}
	for k, v := range util.VersionInfo {
		info[k] = v
	}
	c.JSON(http.StatusOK, info)
}
func (h *Handler) handleDelete(c *gin.Context) {
	givenHmac, err := base64.RawURLEncoding.DecodeString(c.Param("hash"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not decode base64: %v", err)})
		return
	}
	if err := h.store.DeleteEntry(c.Param("id"), givenHmac); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (h *Handler) getURLOrigin(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}
