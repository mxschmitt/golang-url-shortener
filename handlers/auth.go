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
	store := sessions.NewCookieStore([]byte("secret"))

	h.oAuthConf = &oauth2.Config{
		ClientID:     h.config.OAuth.Google.ClientID,
		ClientSecret: h.config.OAuth.Google.ClientSecret,
		RedirectURL:  h.config.BaseURL + "/api/v1/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	h.engine.Use(sessions.Sessions("backend", store))
	h.engine.GET("/api/v1/login", h.handleGoogleLogin)
	h.engine.GET("/api/v1/callback", h.handleGoogleCallback)
	h.engine.POST("/api/v1/check", h.handleGoogleCheck)
}

func (h *Handler) handleGoogleLogin(c *gin.Context) {
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
	err := c.ShouldBind(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := jwt.ParseWithClaims(data.Token, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.config.Secret, nil
	})
	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		c.JSON(http.StatusOK, claims)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

func (h *Handler) handleGoogleCallback(c *gin.Context) {
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	if retrievedState != c.Query("state") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Invalid session state: %s", retrievedState)})
		return
	}

	oAuthToken, err := h.oAuthConf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := h.oAuthConf.Client(oauth2.NoContext, oAuthToken)
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
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 10).Unix(),
		},
		"google",
		user.Sub,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(h.config.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("could not sign token: %v", err)})
		return
	}
	c.HTML(http.StatusOK, "token.tmpl", gin.H{
		"token": tokenString,
	})
}
