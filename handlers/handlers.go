// Package handlers provides the http functionality
package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/store"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	addr   string
	store  store.Store
	engine *gin.Engine
}

// URLUtil is used to help in- and outgoing requests for json
// un- and marshalling
type URLUtil struct {
	URL string
}

// New initializes the http handlers
func New(addr string, store store.Store) *Handler {
	h := &Handler{
		addr:   addr,
		store:  store,
		engine: gin.Default(),
	}
	h.setHandlers()
	return h
}

func (h *Handler) setHandlers() {
	h.engine.POST("/api/v1/create", h.handleCreate)
	h.engine.POST("/api/v1/info", h.handleInfo)
	h.engine.StaticFile("/", "static/index.html")
	h.engine.GET("/:id", h.handleAccess)
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(c *gin.Context) {
	var data struct {
		URL string
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.store.CreateEntry(data.URL, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data.URL = h.getSchemaAndHost(c) + "/" + id
	c.JSON(http.StatusOK, data)
}

func (h *Handler) getSchemaAndHost(c *gin.Context) string {
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s", protocol, c.Request.Host)
}

// handleInfo is the http handler for getting the infos
func (h *Handler) handleInfo(c *gin.Context) {
	var data struct {
		ID string `binding:"required"`
	}
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.store.GetEntryByID(data.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	entry.RemoteAddr = ""
	c.JSON(http.StatusOK, entry)
}

// handleAccess handles the access for incoming requests
func (h *Handler) handleAccess(c *gin.Context) {
	id := c.Param("id")
	entry, err := h.store.GetEntryByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	err = h.store.IncreaseVisitCounter(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, entry.URL)
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(h.addr)
}

// Stop stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
