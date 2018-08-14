// Package handlers provides the http functionality for the URL Shortener
package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/mxschmitt/golang-url-shortener/internal/handlers/tmpls"
	"github.com/mxschmitt/golang-url-shortener/internal/stores"
	"github.com/mxschmitt/golang-url-shortener/internal/util"
	"github.com/pkg/errors"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	store     stores.Store
	engine    *gin.Engine
	providers []string
}

// DoNotPrivateKeyChecking is used for testing
var DoNotPrivateKeyChecking = false

type loggerEntryWithFields interface {
	WithFields(fields logrus.Fields) *logrus.Entry
}

// Ginrus returns a gin.HandlerFunc (middleware) that logs requests using logrus.
//
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
//   3. Optionally, a list of paths to skip logging for (this is why
//      we are not using upstream github.com/gin-gonic/contrib/ginrus)
func Ginrus(logger loggerEntryWithFields, timeFormat string, utc bool, notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		// log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			if utc {
				end = end.UTC()
			}

			entry := logger.WithFields(logrus.Fields{
				"status":     c.Writer.Status(),
				"method":     c.Request.Method,
				"path":       path,
				"ip":         c.ClientIP(),
				"latency":    latency,
				"user-agent": c.Request.UserAgent(),
				"time":       end.Format(timeFormat),
			})

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				entry.Error(c.Errors.String())
			} else {
				entry.Info()
			}
		}
	}
}

// New initializes the http handlers
func New(store stores.Store) (*Handler, error) {
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
	if util.GetConfig().AuthBackend == "oauth" {
		if !DoNotPrivateKeyChecking {
			if err := util.CheckForPrivateKey(); err != nil {
				return nil, errors.Wrap(err, "could not check for private key")
			}
		}
		h.initOAuth()
	} else if util.GetConfig().AuthBackend == "proxy" {
		h.initProxyAuth()
	}
	return h, nil
}

func (h *Handler) addTemplatesFromFS(files []string) error {
	var t *template.Template
	for _, file := range files {
		fileContent, err := tmpls.FSString(false, "/"+file)
		if err != nil {
			return errors.Wrap(err, "could not read template file")
		}
		if t == nil {
			t, err = template.New(file).Parse(fileContent)
			if err != nil {
				return errors.Wrap(err, "could not create template from file content")
			}
			continue
		}
		if _, err := t.New(file).Parse(fileContent); err != nil {
			return errors.Wrap(err, "could not parse template")
		}
	}
	h.engine.SetHTMLTemplate(t)
	return nil
}

func (h *Handler) setHandlers() error {
	if err := h.addTemplatesFromFS([]string{"token.html", "protected.html"}); err != nil {
		return errors.Wrap(err, "could not add templates from FS")
	}
	// only do web access logs if enabled
	if util.GetConfig().EnableAccessLogs {
		if util.GetConfig().EnableDebugMode {
			// in debug mode, log everything including healthchecks
			h.engine.Use(Ginrus(logrus.StandardLogger(), time.RFC3339, false))
		} else {
			// if we are not in debug mode, do not log healthchecks
			h.engine.Use(Ginrus(logrus.StandardLogger(), time.RFC3339, false, "/ok"))
		}
	}
	protected := h.engine.Group("/api/v1/protected")
	switch util.GetConfig().AuthBackend {
	case "oauth":
		logrus.Info("Using OAuth auth backend: oauth")
		protected.Use(h.oAuthMiddleware)
	case "proxy":
		logrus.Info("Using OAuth auth backend: proxy")
		protected.Use(h.proxyAuthMiddleware)
	default:
		logrus.Fatalf("Auth backend method '%s' is not recognized", util.GetConfig().AuthBackend)
	}
	protected.POST("/create", h.handleCreate)
	protected.POST("/lookup", h.handleLookup)
	protected.GET("/recent", h.handleRecent)
	protected.POST("/visitors", h.handleGetVisitors)

	h.engine.GET("/api/v1/info", h.handleInfo)
	h.engine.GET("/d/:id/:hash", h.handleDelete)
	h.engine.GET("/ok", h.handleHealthcheck)
	h.engine.GET("/displayurl", h.handleDisplayURL)

	// Handling the shorted URLs, if no one exists, it checks
	// in the filesystem and sets headers for caching
	h.engine.NoRoute(
		h.handleAccess, // look up shortcuts
		func(c *gin.Context) { // no shortcut found, prep response for FS
			c.Header("Vary", "Accept-Encoding")
			c.Header("Cache-Control", "public, max-age=2592000")
			c.Header("ETag", util.VersionInfo.Commit)
		},
		// Pass down to the embedded FS, but let 404s escape via
		// the interceptHandler.
		gin.WrapH(interceptHandler(http.FileServer(FS(false)), customErrorHandler)),
		// not in FS; redirect to root with customURL target filled out
		func(c *gin.Context) {
			// if we get to this point we should not let the client cache
			c.Header("Cache-Control", "no-cache, no-store")
			c.Redirect(http.StatusTemporaryRedirect, "/?customUrl="+c.Request.URL.Path[1:])
		})
	return nil
}

type interceptResponseWriter struct {
	http.ResponseWriter
	errH func(http.ResponseWriter, int)
}

func (w *interceptResponseWriter) WriteHeader(status int) {
	if status >= http.StatusBadRequest {
		w.errH(w.ResponseWriter, status)
		w.errH = nil
	} else {
		w.ResponseWriter.WriteHeader(status)
	}
}

type errorHandler func(http.ResponseWriter, int)

func (w *interceptResponseWriter) Write(p []byte) (n int, err error) {
	if w.errH == nil {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}

func interceptHandler(next http.Handler, errH errorHandler) http.Handler {
	if errH == nil {
		errH = customErrorHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&interceptResponseWriter{w, errH}, r)
	})
}

func customErrorHandler(w http.ResponseWriter, status int) {
	// let 404s fall through: the next NoRoute handler will redirect
	// them back to the main page with the customURL box filled out.
	if status != 404 {
		http.Error(w, "error", status)
	}
}

// Listen starts the http server
func (h *Handler) Listen() error {
	return h.engine.Run(util.GetConfig().ListenAddr)
}

// CloseStore stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
