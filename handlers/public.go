package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/handlers/auth"
	"github.com/maxibanki/golang-url-shortener/stores/shared"
	"github.com/maxibanki/golang-url-shortener/util"
	"golang.org/x/crypto/bcrypt"
)

// requestHelper is used to help in- and outgoing requests for json
// un- and marshalling
type requestHelper struct {
	URL                       string `binding:"required"`
	ID, DeletionURL, Password string
	Expiration                *time.Time
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
		c.JSON(http.StatusOK, shared.Entry{
			Public: shared.EntryPublicData{
				URL: entry.Public.URL,
			},
		})
		return
	}
	c.JSON(http.StatusOK, entry.Public)
}

// handleAccess handles the access for incoming requests
func (h *Handler) handleAccess(c *gin.Context) {
	id := c.Request.URL.Path[1:]
	entry, err := h.store.GetEntryAndIncrease(id)
	if err != nil {
		if strings.Contains(err.Error(), shared.ErrNoEntryFound.Error()) {
			return
			http.Error(c.Writer, fmt.Sprintf("could not get and crease visitor counter: %v, ", err), http.StatusInternalServerError)
			return
		}
	}
	// No password set
	if len(entry.Password) == 0 {
		c.Redirect(http.StatusTemporaryRedirect, entry.Public.URL)
		go h.registerVisitor(id, c)
		c.Abort()
	} else {
		templateError := ""
		if c.Request.Method == "POST" {
			templateError = func() string {
				pw, exists := c.GetPostForm("password")
				if exists {
					if err := bcrypt.CompareHashAndPassword(entry.Password, []byte(pw)); err != nil {
						return fmt.Sprintf("could not validate password: %v", err)
					}
					return ""
				}
				return "No password set"
			}()
			if templateError == "" {
				c.Redirect(http.StatusSeeOther, entry.Public.URL)
				go h.registerVisitor(id, c)
				c.Abort()
				return
			}
		}
		c.HTML(http.StatusOK, "protected.html", gin.H{
			"ID":    id,
			"Error": templateError,
		})
	}
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(c *gin.Context) {
	var data requestHelper
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := c.MustGet("user").(*auth.JWTClaims)
	id, delID, err := h.store.CreateEntry(shared.Entry{
		Public: shared.EntryPublicData{
			URL:        data.URL,
			Expiration: data.Expiration,
		},
		RemoteAddr:    c.ClientIP(),
		OAuthProvider: user.OAuthProvider,
		OAuthID:       user.OAuthID,
	}, data.ID, data.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	originURL := h.getURLOrigin(c)
	c.JSON(http.StatusOK, requestHelper{
		URL:         fmt.Sprintf("%s/%s", originURL, id),
		DeletionURL: fmt.Sprintf("%s/d/%s/%s", originURL, id, url.QueryEscape(base64.RawURLEncoding.EncodeToString(delID))),
	})
}

// handleGetVisitors handles requests to create an entry
func (h *Handler) handleGetVisitors(c *gin.Context) {
	var data struct {
		ID string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dataSets, err := h.store.GetVisitors(data.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dataSets)
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

func (h *Handler) handleRecent(c *gin.Context) {
	user := c.MustGet("user").(*auth.JWTClaims)
	entries, err := h.store.GetUserEntries(user.OAuthProvider, user.OAuthID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for k, entry := range entries {
		mac := hmac.New(sha512.New, util.GetPrivateKey())
		if _, err := mac.Write([]byte(k)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		entry.DeletionURL = fmt.Sprintf("%s/d/%s/%s", h.getURLOrigin(c), k, url.QueryEscape(base64.RawURLEncoding.EncodeToString(mac.Sum(nil))))
		entries[k] = entry
	}
	c.JSON(http.StatusOK, entries)
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
	if c.Request.TLS != nil || util.GetConfig().UseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}

func (h *Handler) registerVisitor(id string, c *gin.Context) {
	h.store.RegisterVisit(id, shared.Visitor{
		IP:          c.ClientIP(),
		Timestamp:   time.Now(),
		Referer:     c.GetHeader("Referer"),
		UserAgent:   c.GetHeader("User-Agent"),
		UTMSource:   c.Query("utm_source"),
		UTMMedium:   c.Query("utm_medium"),
		UTMCampaign: c.Query("utm_campaign"),
		UTMContent:  c.Query("utm_content"),
		UTMTerm:     c.Query("utm_term"),
	})
}
