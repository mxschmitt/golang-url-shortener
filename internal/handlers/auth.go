package handlers

import (
	"fmt"
	"net/http"

	"github.com/mxschmitt/golang-url-shortener/internal/handlers/auth"
	"github.com/mxschmitt/golang-url-shortener/internal/util"
	"github.com/sirupsen/logrus"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (h *Handler) initOAuth() {
	switch backend := util.GetConfig().Backend; backend {
	// use redis as the session store if it is configured
	case "redis":
		store, _ := redis.NewStoreWithDB(10, "tcp", util.GetConfig().Redis.Host, util.GetConfig().Redis.Password, util.GetConfig().Redis.SessionDB, util.GetPrivateKey())
		h.engine.Use(sessions.Sessions("backend", store))
	default:
		h.engine.Use(sessions.Sessions("backend", cookie.NewStore(util.GetPrivateKey())))
	}
	h.providers = []string{}
	google := util.GetConfig().Google
	if google.Enabled() {
		auth.WithAdapterWrapper(auth.NewGoogleAdapter(google.ClientID, google.ClientSecret), h.engine.Group("/api/v1/auth/google"))
		h.providers = append(h.providers, "google")
	}
	github := util.GetConfig().GitHub
	if github.Enabled() {
		auth.WithAdapterWrapper(auth.NewGithubAdapter(github.ClientID, github.ClientSecret, github.EndpointURL), h.engine.Group("/api/v1/auth/github"))
		h.providers = append(h.providers, "github")
	}
	microsoft := util.GetConfig().Microsoft
	if microsoft.Enabled() {
		auth.WithAdapterWrapper(auth.NewMicrosoftAdapter(microsoft.ClientID, microsoft.ClientSecret), h.engine.Group("/api/v1/auth/microsoft"))
		h.providers = append(h.providers, "microsoft")
	}
	okta := util.GetConfig().Okta
	if okta.Enabled() {
		auth.WithAdapterWrapper(auth.NewOktaAdapter(okta.ClientID, okta.ClientSecret, okta.EndpointURL), h.engine.Group("/api/v1/auth/okta"))
		h.providers = append(h.providers, "okta")
	}

	genericOIDC := util.GetConfig().GenericOIDC
	if genericOIDC.Enabled() {
		auth.WithAdapterWrapper(auth.NewGenericOIDCAdapter(genericOIDC.ClientID, genericOIDC.ClientSecret, genericOIDC.EndpointURL), h.engine.Group("/api/v1/auth/generic_oidc"))
		h.providers = append(h.providers, "generic_oidc")
	}

	h.engine.POST("/api/v1/auth/check", h.handleAuthCheck)
}

// initProxyAuth intializes data structures for proxy authentication mode
func (h *Handler) initProxyAuth() {
	switch backend := util.GetConfig().Backend; backend {
	// use redis as the session store if it is configured
	case "redis":
		store, _ := redis.NewStoreWithDB(10, "tcp", util.GetConfig().Redis.Host, util.GetConfig().Redis.Password, util.GetConfig().Redis.SessionDB, util.GetPrivateKey())
		h.engine.Use(sessions.Sessions("backend", store))
	default:
		h.engine.Use(sessions.Sessions("backend", cookie.NewStore(util.GetPrivateKey())))
	}
	h.providers = []string{}
	h.providers = append(h.providers, "proxy")
	h.engine.POST("/api/v1/auth/check", h.handleAuthCheck)
}

func (h *Handler) parseJWT(wt string) (*auth.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(wt, &auth.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return util.GetPrivateKey(), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	return token.Claims.(*auth.JWTClaims), nil
}

// oAuthMiddleware implements an auth layer that validates a JWT token
func (h *Handler) oAuthMiddleware(c *gin.Context) {
	authError := func() error {
		wt := c.GetHeader("Authorization")
		if wt == "" {
			return errors.New("Authorization header not set")
		}
		claims, err := h.parseJWT(wt)
		if err != nil {
			return errors.Wrap(err, "could not parse JWT")
		}
		c.Set("user", claims)
		return nil
	}()
	if authError != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "authentication failed",
		})
		logrus.Debugf("Authentication middleware check failed: %v\n", authError)
		return
	}
	c.Next()
}

// proxyAuthMiddleware implements an auth layer that trusts (and
// optionally requires) header data from an identity-aware proxy
func (h *Handler) proxyAuthMiddleware(c *gin.Context) {
	authError := func() error {
		claims, err := h.fakeClaimsForProxy(c)
		if err != nil {
			return err
		}
		c.Set("user", claims)
		return nil
	}()
	if authError != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "authentication failed",
		})
		logrus.Errorf("Authentication middleware check failed: %v\n", authError)
		return
	}
	c.Next()
}

// fakeClaimsForProxy returns a pointer to a auth.JWTClaims struct containing
// data pulled from headers inserted by an identity-aware proxy.
func (h *Handler) fakeClaimsForProxy(c *gin.Context) (*auth.JWTClaims, error) {
	uid := c.GetHeader(util.GetConfig().Proxy.UserHeader)
	logrus.Debugf("Got proxy uid '%s' from header '%s'", uid, util.GetConfig().Proxy.UserHeader)
	if uid == "" {
		logrus.Debugf("No proxy uid found!")
		if util.GetConfig().Proxy.RequireUserHeader {
			msg := fmt.Sprintf("Required authorization header not set: %s", util.GetConfig().Proxy.UserHeader)
			logrus.Error(msg)
			return nil, errors.New(msg)
		}
		logrus.Debugf("Setting uid to 'anonymous'")
		uid = "anonymous"
	}
	// optionally pick a display name out of the headers as well; if we
	// can't find it, just use the uid.
	displayName := c.GetHeader(util.GetConfig().Proxy.DisplayNameHeader)
	logrus.Debugf("Got proxy display name '%s' from header '%s'", displayName, util.GetConfig().Proxy.DisplayNameHeader)
	if displayName == "" {
		logrus.Debugf("Setting displayname to '%s'", uid)
		displayName = uid
	}
	// it's not actually oauth but the naming convention is too
	// deeply embedded in the code for it to be worth changing.
	claims := &auth.JWTClaims{
		OAuthID:       uid,
		OAuthName:     displayName,
		OAuthPicture:  "/images/proxy_user.png",
		OAuthProvider: "proxy",
	}
	return claims, nil
}

func (h *Handler) handleAuthCheck(c *gin.Context) {
	var data struct {
		Token string `binding:"required"`
	}
	var claims *auth.JWTClaims
	var err error

	if err = c.ShouldBind(&data); err != nil {
		logrus.Errorf("Did not bind correctly: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if util.GetConfig().AuthBackend == "proxy" {
		// for proxy auth, we trust that the proxy has taken care of things
		// for us and we are only testing that the middleware successfully
		// pulled the necessary headers from the request.
		claims, err = h.fakeClaimsForProxy(c)
	} else {
		claims, err = h.parseJWT(data.Token)
	}
	if err != nil {
		logrus.Errorf("Could not parse auth data: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	sessionData := gin.H{
		"ID":       claims.OAuthID,
		"Name":     claims.OAuthName,
		"Picture":  claims.OAuthPicture,
		"Provider": claims.OAuthProvider,
	}
	logrus.Debugf("Found session data: %v", sessionData)
	c.JSON(http.StatusOK, sessionData)
}

func (h *Handler) oAuthPropertiesEquals(c *gin.Context, oauthID, oauthProvider string) bool {
	user := c.MustGet("user").(*auth.JWTClaims)
	if oauthID == user.OAuthID && oauthProvider == user.OAuthProvider {
		return true
	}
	return false
}
