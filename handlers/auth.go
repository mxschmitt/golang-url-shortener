package handlers

import (
	"net/http"

	"github.com/mxschmitt/golang-url-shortener/handlers/auth"
	"github.com/mxschmitt/golang-url-shortener/util"
	"github.com/sirupsen/logrus"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (h *Handler) initOAuth() {
	h.engine.Use(sessions.Sessions("backend", sessions.NewCookieStore(util.GetPrivateKey())))

	h.providers = []string{}
	google := util.GetConfig().Google
	if google.Enabled() {
		auth.WithAdapterWrapper(auth.NewGoogleAdapter(google.ClientID, google.ClientSecret), h.engine.Group("/api/v1/auth/google"))
		h.providers = append(h.providers, "google")
	}
	github := util.GetConfig().GitHub
	if github.Enabled() {
		auth.WithAdapterWrapper(auth.NewGithubAdapter(github.ClientID, github.ClientSecret), h.engine.Group("/api/v1/auth/github"))
		h.providers = append(h.providers, "github")
	}
	microsoft := util.GetConfig().Microsoft
	if microsoft.Enabled() {
		auth.WithAdapterWrapper(auth.NewMicrosoftAdapter(microsoft.ClientID, microsoft.ClientSecret), h.engine.Group("/api/v1/auth/microsoft"))
		h.providers = append(h.providers, "microsoft")
	}

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

func (h *Handler) authMiddleware(c *gin.Context) {
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

func (h *Handler) handleAuthCheck(c *gin.Context) {
	var data struct {
		Token string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := h.parseJWT(data.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ID":       claims.OAuthID,
		"Name":     claims.OAuthName,
		"Picture":  claims.OAuthPicture,
		"Provider": claims.OAuthProvider,
	})
}

func (h *Handler) oAuthPropertiesEquals(c *gin.Context, oauthID, oauthProvider string) bool {
	user := c.MustGet("user").(*auth.JWTClaims)
	if oauthID == user.OAuthID && oauthProvider == user.OAuthProvider {
		return true
	}
	return false
}
