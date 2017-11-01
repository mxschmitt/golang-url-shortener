// Package handlers provides the http functionality
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/store"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Handler holds the funcs and attributes for the
// http communication
type Handler struct {
	config    config.Handlers
	store     store.Store
	engine    *gin.Engine
	oAuthConf *oauth2.Config
}

// URLUtil is used to help in- and outgoing requests for json
// un- and marshalling
type URLUtil struct {
	URL string
}

type oAuthUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
	Hd            string `json:"hd"`
}

// New initializes the http handlers
func New(handlerConfig config.Handlers, store store.Store) *Handler {
	h := &Handler{
		config: handlerConfig,
		store:  store,
		engine: gin.Default(),
	}
	h.setHandlers()
	h.initOAuth()
	return h
}

func (h *Handler) setHandlers() {
	if !h.config.EnableGinDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	h.engine.POST("/api/v1/create", h.handleCreate)
	h.engine.POST("/api/v1/info", h.handleInfo)
	// h.engine.Static("/static", "static/src")
	h.engine.NoRoute(h.handleAccess)
}

func (h *Handler) initOAuth() {
	store := sessions.NewCookieStore([]byte("secret"))

	h.oAuthConf = &oauth2.Config{
		ClientID:     h.config.OAuth.Google.ClientID,
		ClientSecret: h.config.OAuth.Google.ClientSecret,
		RedirectURL:  "http://127.0.0.1:3000/api/v1/auth/",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	h.engine.Use(sessions.Sessions("goquestsession", store))
	h.engine.GET("/api/v1/login", h.handleGoogleLogin)

	private := h.engine.Group("/api/v1/auth")
	private.Use(h.handleGoogleAuth)
	private.GET("/", h.handleGoogleCallback)
	private.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from private for groups"})
	})
}

func (h *Handler) randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (h *Handler) handleGoogleAuth(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Errorf("Invalid session state: %s", retrievedState)})
		return
	}

	token, err := h.oAuthConf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := h.oAuthConf.Client(oauth2.NoContext, token)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer userinfo.Body.Close()
	data, err := ioutil.ReadAll(userinfo.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not read body: %v", err)})
		return
	}

	var user oAuthUser
	err = json.Unmarshal(data, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Decoding userinfo failed: %v", err)})
		return
	}
	c.Set("user", user)
}

func (h *Handler) handleGoogleLogin(c *gin.Context) {
	state := h.randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, h.oAuthConf.AuthCodeURL(state))
}

func (h *Handler) handleGoogleCallback(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"Hello": "from private", "user": ctx.MustGet("user").(oAuthUser)})
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
	var id string
	if len(c.Request.URL.Path) > 1 {
		id = c.Request.URL.Path[1:]
	}
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
	return h.engine.Run(h.config.ListenAddr)
}

// CloseStore stops the http server and the closes the db gracefully
func (h *Handler) CloseStore() error {
	return h.store.Close()
}
