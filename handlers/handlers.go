// Package handlers provides the http functionality for the URL Shortener
package handlers

import (
	"crypto/rand"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/ginrus"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/handlers/tmpls"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	config                 config.Handlers
	store                  store.Store
	engine                 *gin.Engine
	oAuthConf              *oauth2.Config
	log                    *logrus.Logger
	DoNotCheckConfigViaGet bool // DoNotCheckConfigViaGet is for the unit testing usage
}

// New initializes the http handlers
func New(handlerConfig config.Handlers, store store.Store, log *logrus.Logger, testing bool) (*Handler, error) {
	if !handlerConfig.EnableDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	h := &Handler{
		config: handlerConfig,
		store:  store,
		log:    log,
		engine: gin.New(),
	}
	if err := h.setHandlers(); err != nil {
		return nil, errors.Wrap(err, "could not set handlers")
	}
	if !testing {
		if err := h.checkIfSecretExist(); err != nil {
			return nil, errors.Wrap(err, "could not check if secret exist")
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
	templ, err := template.New(name).Funcs(h.engine.FuncMap).Parse(tokenTemplate)
	if err != nil {
		return errors.Wrap(err, "could not create template from file content")
	}
	h.engine.SetHTMLTemplate(templ)
	return nil
}

func (h *Handler) checkIfSecretExist() error {
	if !h.DoNotCheckConfigViaGet {
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
	}
	return nil
}

func (h *Handler) setHandlers() error {
	if err := h.setTemplateFromFS("token.tmpl"); err != nil {
		return errors.Wrap(err, "could not set template from FS")
	}
	h.engine.Use(ginrus.Ginrus(h.log, time.RFC3339, false))
	protected := h.engine.Group("/api/v1/protected")
	protected.Use(h.authMiddleware)
	protected.POST("/create", h.handleCreate)
	protected.POST("/lookup", h.handleLookup)

	h.engine.NoRoute(h.handleAccess, gin.WrapH(http.FileServer(FS(false))))
	return nil
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(h.config.ListenAddr)
}

// CloseStore stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
