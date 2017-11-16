package handlers

import (
	"net/http"

	"github.com/maxibanki/golang-url-shortener/handlers/auth"
	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (h *Handler) initOAuth() {
	h.engine.Use(sessions.Sessions("backend", sessions.NewCookieStore(util.GetPrivateKey())))

	auth.WithAdapterWrapper(auth.NewGoogleAdapter(viper.GetString("Google.ClientID"), viper.GetString("Google.ClientSecret"), viper.GetString("base_url")), h.engine.Group("/api/v1/auth/google"))
	auth.WithAdapterWrapper(auth.NewGithubAdapter(viper.GetString("GitHub.ClientID"), viper.GetString("GitHub.ClientSecret"), viper.GetString("base_url")), h.engine.Group("/api/v1/auth/github"))

	h.engine.POST("/api/v1/check", h.handleAuthCheck)
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
			return errors.New("'Authorization' header not set")
		}
		claims, err := h.parseJWT(wt)
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
		logrus.Debugf("Authentication middleware failed: %v\n", authError)
		return
	}
	c.Next()
}

func (h *Handler) handleAuthCheck(c *gin.Context) {
	var data struct {
		Token string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, err := h.parseJWT(data.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ID":       claims.OAuthID,
		"Name":     claims.OAuthName,
		"Picture":  claims.OAuthPicture,
		"Provider": claims.OAuthProvider,
	})
}
