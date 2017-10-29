package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/maxibanki/golang-url-shorter/store"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	addr   string
	store  store.Store
	server *http.Server
}

// URLUtil is used to help in- and outgoing requests for json
// un- and marshalling
type URLUtil struct {
	URL string
}

// New initializes the http handlers
func New(addr string, store store.Store) *Handler {
	h := &Handler{
		addr:  addr,
		store: store,
	}
	router := h.handlers()
	h.server = &http.Server{Addr: h.addr, Handler: router}
	return h
}

func (h *Handler) handlers() *httprouter.Router {
	router := httprouter.New()
	router.POST("/api/v1/create", h.handleCreate)
	router.POST("/api/v1/info", h.handleInfo)
	router.GET("/:id", h.handleAccess)
	return router
}

// handleCreate handles requests to create an entry
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		h.handleCreateJSON(w, r)
		break
	case "application/x-www-form-urlencoded":
		h.handleCreateForm(w, r)
		break
	default:
		if strings.Contains(contentType, "multipart/form-data;") {
			h.handleCreateMultipartForm(w, r)
			return
		}
	}
}

func (h *Handler) handleCreateJSON(w http.ResponseWriter, r *http.Request) {
	var req URLUtil
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode JSON: %v", err), http.StatusBadRequest)
		return
	}
	id, err := h.store.CreateEntry(req.URL, r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.URL = h.generateURL(r, id)
	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) handleCreateMultipartForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1048576)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, ok := r.MultipartForm.Value["URL"]; !ok {
		http.Error(w, "URL key does not exist", http.StatusBadRequest)
		return
	}
	id, err := h.store.CreateEntry(r.MultipartForm.Value["URL"][0], r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req URLUtil
	req.URL = h.generateURL(r, id)
	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) handleCreateForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.PostFormValue("URL") == "" {
		http.Error(w, "URL key does not exist", http.StatusBadRequest)
		return
	}
	id, err := h.store.CreateEntry(r.PostFormValue("URL"), r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req URLUtil
	req.URL = h.generateURL(r, id)
	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handler) generateURL(r *http.Request, id string) string {
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s", protocol, r.Host, id)
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
	raw, err := h.store.GetEntryByIDRaw(req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)

		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(raw)
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
	return h.server.ListenAndServe()
}

// Stop stops the http server and the closes the db gracefully
func (h *Handler) Stop() error {
	err := h.store.Close()
	if err != nil {
		return err
	}
	return h.server.Shutdown(context.Background())
}
