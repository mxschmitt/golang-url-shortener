// Package handlers provides the http functionality
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/julienschmidt/httprouter"
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
	// h.engine.POST("/api/v1/info", h.handleInfo)
	// h.engine.GET("/:id", h.handleAccess)
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

	id, err := h.store.CreateEntry(data.URL, c.Request.RemoteAddr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}
	data.URL = fmt.Sprintf("%s://%s/%s", protocol, c.Request.Host, id)
	c.JSON(http.StatusOK, data)
}

// handleInfo is the http handler for getting the infos
func (h *Handler) handleInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req struct {
		ID string
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode JSON: %v", err), http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		http.Error(w, "no ID provided", http.StatusBadRequest)
		return
	}
	entry, err := h.store.GetEntryByID(req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	entry.RemoteAddr = ""
	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

// handleAccess handles the access for incoming requests
func (h *Handler) handleAccess(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	entry, err := h.store.GetEntryByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = h.store.IncreaseVisitCounter(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Redirect(w, r, entry.URL, http.StatusTemporaryRedirect)
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(h.addr)
}

// Stop stops the http server and the closes the db gracefully
func (h *Handler) Stop() error {
	// err := h.store.Close()
	// if err != nil {
	// 	return err
	// }
	return h.store.Close()
	// return h.server.Shutdown(context.Background())
}
