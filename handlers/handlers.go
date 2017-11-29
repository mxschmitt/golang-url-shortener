// Package handlers provides the http functionality for the URL Shortener
package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/handlers/tmpls"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/pkg/errors"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	store     store.Store
	engine    *gin.Engine
	providers []string
}

// DoNotPrivateKeyChecking is used for testing
var DoNotPrivateKeyChecking = false

// New initializes the http handlers
func New(store store.Store) (*Handler, error) {
	if !util.GetConfig().EnableDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	h := &Handler{
		store:  store,
		engine: gin.New(),
	}
	if err := h.setHandlers(); err != nil {
		return nil, errors.Wrap(err, "could not set handlers")
	}
	if !DoNotPrivateKeyChecking {
		if err := util.CheckForPrivateKey(); err != nil {
			return nil, errors.Wrap(err, "could not check for private key")
		}
	}
	h.initOAuth()
	return h, nil
}

func (h *Handler) setTemplateFromFS(name string) error {
	tokenTemplate, err := tmpls.FSString(false, "/"+name)
	if err != nil {
		return errors.Wrap(err, "could not read token template file")
	}
	templ, err := template.New(name).Parse(tokenTemplate)
	if err != nil {
		return errors.Wrap(err, "could not create template from file content")
	}
	h.engine.SetHTMLTemplate(templ)
	return nil
}

func (h *Handler) setHandlers() error {
	templates := []string{"token.html", "protected.html"}
	for _, template := range templates {
		if err := h.setTemplateFromFS(template); err != nil {
			return errors.Wrapf(err, "could not set template %s from FS", template)
		}
	}
	h.engine.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, false))
	protected := h.engine.Group("/api/v1/protected")
	protected.Use(h.authMiddleware)
	protected.POST("/create", h.handleCreate)
	protected.POST("/lookup", h.handleLookup)
	protected.POST("/recent", h.handleRecent)
	protected.POST("/visitors", h.handleGetVisitors)

	h.engine.GET("/api/v1/info", h.handleInfo)
	h.engine.GET("/d/:id/:hash", h.handleDelete)

	// Handling the shorted URLs, if no one exists, it checks
	// in the filesystem and sets headers for caching
	h.engine.NoRoute(h.handleAccess, func(c *gin.Context) {
		c.Header("Vary", "Accept-Encoding")
		c.Header("Cache-Control", "public, max-age=2592000")
		c.Header("ETag", util.VersionInfo["commit"])
	}, gin.WrapH(http.FileServer(FS(false))))
	return nil
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(util.GetConfig().ListenAddr)
}

// CloseStore stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
