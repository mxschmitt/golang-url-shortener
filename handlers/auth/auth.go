package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/maxibanki/golang-url-shortener/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Adapter will be implemented by each oAuth provider
type Adapter interface {
	GetRedirectURL(state string) string
	GetUserData(state, code string) (*user, error)
	GetOAuthProviderName() string
}

type user struct {
	ID, Name, Picture string
}

// JWTClaims are the data and general information which is stored in the JWT
type JWTClaims struct {
	jwt.StandardClaims
	OAuthProvider string
	OAuthID       string
	OAuthName     string
	OAuthPicture  string
}

// AdapterWrapper wraps an normal oAuth Adapter with some generic functions
// to be implemented directly by the gin router
type AdapterWrapper struct{ Adapter }

// WithAdapterWrapper creates an adapterWrapper out of the oAuth Adapter and an gin.RouterGroup
func WithAdapterWrapper(a Adapter, h *gin.RouterGroup) *AdapterWrapper {
	aw := &AdapterWrapper{a}
	h.GET("/login", aw.HandleLogin)
	h.GET("/callback", aw.HandleCallback)
	return aw
}

// HandleLogin handles the incoming http request for the oAuth process
// and redirects to the generated URL of the provider
func (a *AdapterWrapper) HandleLogin(c *gin.Context) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(c.Writer, fmt.Sprintf("could not read random state: %v", err), http.StatusInternalServerError)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(b)
	session := sessions.Default(c)
	session.Set("state", state)
	if err := session.Save(); err != nil {
		http.Error(c.Writer, fmt.Sprintf("could not save state to session: %v", err), http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, a.GetRedirectURL(state))
}

// HandleCallback handles and validates the callback which is coming back from the oAuth request
func (a *AdapterWrapper) HandleCallback(c *gin.Context) {
	session := sessions.Default(c)
	sessionState := session.Get("state")
	receivedState := c.Query("state")
	if sessionState != receivedState {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid session state: %s", sessionState)})
		return
	}
	user, err := a.GetUserData(receivedState, c.Query("code"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	logrus.WithFields(logrus.Fields{
		"Provider": a.GetOAuthProviderName(),
		"Name":     user.Name,
	}).Info("New user logged in via oAuth")
	token, err := a.newJWT(user, a.GetOAuthProviderName())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "token.html", gin.H{
		"token": token,
	})
}

func (a *AdapterWrapper) newJWT(user *user, provider string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
		},
		provider,
		user.ID,
		user.Name,
		user.Picture,
	})
	tokenString, err := token.SignedString(util.GetPrivateKey())
	if err != nil {
		return "", errors.Wrap(err, "could not sign token")
	}
	return tokenString, nil
}
