// Package handlers provides the http functionality
package handlers

import (
	"crypto/rand"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	config    config.Handlers
	store     store.Store
	engine    *gin.Engine
	oAuthConf *oauth2.Config
}

// New initializes the http handlers
func New(handlerConfig config.Handlers, store store.Store) (*Handler, error) {
	h := &Handler{
		config: handlerConfig,
		store:  store,
		engine: gin.Default(),
	}
	h.setHandlers()
	if err := h.checkIfSecretExist(); err != nil {
		return nil, errors.Wrap(err, "could not check if secret exist")
	}
	h.initOAuth()
	return h, nil
}

func (h *Handler) checkIfSecretExist() error {
	conf := config.Get()
	if conf.Handlers.Secret == nil {
		b := make([]byte, 128)
		if _, err := rand.Read(b); err != nil {
			return err
		}
		conf.Handlers.Secret = b
		if err := config.Set(conf); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) setHandlers() {
	if !h.config.EnableGinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	h.engine.POST("/api/v1/create", h.handleCreate)
	h.engine.POST("/api/v1/info", h.handleInfo)
	// h.engine.Static("/static", "static/src")
	h.engine.NoRoute(h.handleAccess)
	h.engine.LoadHTMLGlob("templates/*")
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(h.config.ListenAddr)
}

// CloseStore stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
