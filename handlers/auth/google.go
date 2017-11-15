package auth

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleAdapter struct {
	config *oauth2.Config
}

func NewGoogleAdapter(clientID, clientSecret, baseURL string) Adapter {
	return &googleAdapter{&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/v1/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}}
}

func (a *googleAdapter) GetRedirectURl(state string) string {
	return a.config.AuthCodeURL(state)
}

func (a *googleAdapter) GetUserData(state, code string) (*user, error) {
	oAuthToken, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.Wrap(err, "could not exchange code")
	}

	client := a.config.Client(context.Background(), oAuthToken)
	oAuthUserInfoReq, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, errors.Wrap(err, "could not get user data")
	}
	defer oAuthUserInfoReq.Body.Close()
	data, err := ioutil.ReadAll(oAuthUserInfoReq.Body)
	if err != nil {
		return nil, errors.Wrap(err, "could not read body")
	}
	var gUser struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err = json.Unmarshal(data, &gUser); err != nil {
		return nil, errors.Wrap(err, "decoding user info failed")
	}
	return &user{
		ID:      gUser.Sub,
		Name:    gUser.Name,
		Picture: gUser.Picture,
	}, nil
}

func (a *googleAdapter) GetOAuthProviderName() string {
	return "google"
}
