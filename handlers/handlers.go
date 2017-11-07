// Package handlers provides the http functionality
//go:generate esc -o static.go -pkg handlers -prefix ../static/build ../static/build
//go:generate esc -o tmpls/tmpls.go -pkg tmpls -include ^*\.tmpl -prefix tmpls tmpls
package handlers

import (
	"crypto/rand"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/handlers/tmpls"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	config                 config.Handlers
	store                  store.Store
	engine                 *gin.Engine
	oAuthConf              *oauth2.Config
	DoNotCheckConfigViaGet bool // DoNotCheckConfigViaGet is for the unit testing usage
}

// New initializes the http handlers
func New(handlerConfig config.Handlers, store store.Store) (*Handler, error) {
	h := &Handler{
		config: handlerConfig,
		store:  store,
		engine: gin.Default(),
	}
	if err := h.setHandlers(); err != nil {
		return nil, errors.Wrap(err, "could not set handlers")
	}
	if err := h.checkIfSecretExist(); err != nil {
		return nil, errors.Wrap(err, "could not check if secret exist")
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
	if h.DoNotCheckConfigViaGet {
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
	if !h.config.EnableGinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	protected := h.engine.Group("/api/v1/protected")
	protected.Use(h.authMiddleware)
	protected.POST("/create", h.handleCreate)
	protected.POST("/info", h.handleInfo)

	h.engine.NoRoute(h.handleAccess, gin.WrapH(http.FileServer(FS(false))))

	if err := h.setTemplateFromFS("token.tmpl"); err != nil {
		return errors.Wrap(err, "could not set template from FS")
	}
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
