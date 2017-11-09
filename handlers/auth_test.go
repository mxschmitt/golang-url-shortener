package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/maxibanki/golang-url-shortener/config"
	"github.com/maxibanki/golang-url-shortener/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
)

const (
	testingDBName = "main.db"
)

var (
	secret           = []byte("our really great secret")
	server           *httptest.Server
	closeServer      func() error
	handler          *Handler
	testingClaimData = jwtClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365).Unix(),
		},
		"google",
		"sub sub sub",
		"name",
		"url",
	}
	tokenString string
)

func TestCreateBackend(t *testing.T) {
	store, err := store.New(config.Store{
		DBPath:          testingDBName,
		ShortedIDLength: 4,
	})
	if err != nil {
		t.Fatalf("could not create store: %v", err)
	}
	handler, err := New(config.Handlers{
		ListenAddr: ":8080",
		Secret:     secret,
		BaseURL:    "http://127.0.0.1",
	}, *store, logrus.New())
	if err != nil {
		t.Fatalf("could not create handler: %v", err)
	}
	handler.DoNotCheckConfigViaGet = true
	server = httptest.NewServer(handler.engine)
	closeServer = func() error {
		server.Close()
		if err := handler.CloseStore(); err != nil {
			return errors.Wrap(err, "could not close store")
		}
		if err := os.Remove(testingDBName); err != nil {
			return errors.Wrap(err, "could not remove testing db")
		}
		return nil
	}
}

func TestHandleGoogleRedirect(t *testing.T) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}, // don't follow redirects
	}
	resp, err := client.Get(server.URL + "/api/v1/login")
	if err != nil {
		t.Fatalf("could not get login request: %v", err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status code: %d; got: %d", http.StatusTemporaryRedirect, resp.StatusCode)
	}
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, google.Endpoint.AuthURL) {
		t.Fatalf("redirect is not correct, got: %s", location)
	}
}

func TestCreateNewJWT(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testingClaimData)
	var err error
	tokenString, err = token.SignedString(secret)
	if err != nil {
		t.Fatalf("could not sign token: %v", err)
	}
}

func TestForbiddenReqest(t *testing.T) {
	resp, err := http.Post(server.URL+"/api/v1/protected/create", "application/json", nil)
	if err != nil {
		t.Fatalf("could not execute get request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("incorrect status code: %d; got: %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestInvalidToken(t *testing.T) {
	req, err := http.NewRequest("POST", server.URL+"/api/v1/protected/create", nil)
	if err != nil {
		t.Fatalf("could not create request %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "incorrect one")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("could not execute post request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("incorrect status code: %d; got: %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestCheckToken(t *testing.T) {
	body, err := json.Marshal(map[string]string{
		"Token": tokenString,
	})
	if err != nil {
		t.Fatalf("could not post to the backend: %v", err)
	}
	resp, err := http.Post(server.URL+"/api/v1/check", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("could not execute get request: %v", err)
	}
	var data checkResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}
	tt := []struct {
		name          string
		currentValue  string
		expectedValue string
	}{
		{
			name:          "ID",
			currentValue:  data.ID,
			expectedValue: testingClaimData.OAuthID,
		},
		{
			name:          "Name",
			currentValue:  data.Name,
			expectedValue: testingClaimData.OAuthName,
		},
		{
			name:          "Picture",
			currentValue:  data.Picture,
			expectedValue: testingClaimData.OAuthPicture,
		},
		{
			name:          "Provider",
			currentValue:  data.Provider,
			expectedValue: testingClaimData.OAuthProvider,
		},
	}
	for _, tc := range tt {
		t.Run(fmt.Sprintf("Checking: %s", tc.name), func(t *testing.T) {
			if tc.currentValue != tc.expectedValue {
				t.Fatalf("incorrect jwt value: %s; expected: %s", tc.expectedValue, tc.currentValue)
			}
		})
	}

}
func TestCloseBackend(t *testing.T) {
	if err := closeServer(); err != nil {
		t.Fatalf("could not close server: %v", err)
	}
}
