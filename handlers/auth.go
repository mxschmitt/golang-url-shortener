package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type jwtClaims struct {
	jwt.StandardClaims
	OAuthProvider string
	OAuthID       string
	OAuthName     string
	OAuthPicture  string
	OAuthEmail    string
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

func (h *Handler) initOAuth() {
	h.oAuthConf = &oauth2.Config{
		ClientID:     h.config.OAuth.Google.ClientID,
		ClientSecret: h.config.OAuth.Google.ClientSecret,
		RedirectURL:  h.config.BaseURL + "/api/v1/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	h.engine.Use(sessions.Sessions("backend", sessions.NewCookieStore(h.config.Secret)))
	h.engine.GET("/api/v1/login", h.handleGoogleRedirect)
	h.engine.GET("/api/v1/callback", h.handleGoogleCallback)
	h.engine.POST("/api/v1/check", h.handleGoogleCheck)
}

func (h *Handler) handleGoogleRedirect(c *gin.Context) {
	state := h.randToken()
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, h.oAuthConf.AuthCodeURL(state))
}

func (h *Handler) handleGoogleCheck(c *gin.Context) {
	var data struct {
		Token string `binding:"required"`
	}
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := jwt.ParseWithClaims(data.Token, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.config.Secret, nil
	})
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		c.JSON(http.StatusOK, gin.H{
			"ID":       claims.OAuthID,
			"Email":    claims.OAuthEmail,
			"Name":     claims.OAuthName,
			"Picture":  claims.OAuthPicture,
			"Provider": claims.OAuthProvider,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func (h *Handler) handleGoogleCallback(c *gin.Context) {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid session state: %s", retrievedState)})
		return
	}

	oAuthToken, err := h.oAuthConf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not exchange code: %v", err)})
		return
	}

	client := h.oAuthConf.Client(oauth2.NoContext, oAuthToken)
	oAuthUserInfoReq, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("could not get user data: %v", err)})
		return
	}
	defer oAuthUserInfoReq.Body.Close()
	data, err := ioutil.ReadAll(oAuthUserInfoReq.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not read body: %v", err)})
		return
	}
	var user oAuthUser
	if err = json.Unmarshal(data, &user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("decoding user info failed: %v", err)})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
		},
		"google",
		user.Sub,
		user.Name,
		user.Picture,
		user.Email,
	})

	tokenString, err := token.SignedString(h.config.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not sign token: %v", err)})
		return
	}
	c.HTML(http.StatusOK, "token.tmpl", gin.H{
		"token": tokenString,
	})
}
