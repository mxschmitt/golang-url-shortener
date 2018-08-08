package auth

import (
	"context"
	"encoding/json"

	"github.com/mxschmitt/golang-url-shortener/internal/util"
	"github.com/sirupsen/logrus"

	"golang.org/x/oauth2/github"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type githubAdapter struct {
	config *oauth2.Config
}

// NewGithubAdapter creates an oAuth adapter out of the credentials and the baseURL
func NewGithubAdapter(clientID, clientSecret string) Adapter {
	if util.GetConfig().GitHubEndpointURL != "" {
		github.Endpoint.AuthURL = util.GetConfig().GitHubEndpointURL + "/login/oauth/authorize"
		github.Endpoint.TokenURL = util.GetConfig().GitHubEndpointURL + "/login/oauth/access_token"
	}
	return &githubAdapter{&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  util.GetConfig().BaseURL + "/api/v1/auth/github/callback",
		Scopes: []string{
			"(no scope)",
		},
		Endpoint: github.Endpoint,
	}}
}

func (a *githubAdapter) GetRedirectURL(state string) string {
	return a.config.AuthCodeURL(state)
}

func (a *githubAdapter) GetUserData(state, code string) (*user, error) {
	logrus.Debugf("Getting User Data with state: %s, and code: %s", state, code)
	oAuthToken, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.Wrap(err, "could not exchange code")
	}

	gitHubUserURL := "https://github.homedepot.com/api/v3/user"
	if util.GetConfig().GitHubEndpointURL != "" {
		gitHubUserURL = util.GetConfig().GitHubEndpointURL + "/api/v3/user"
	}
	oAuthUserInfoReq, err := a.config.Client(context.Background(), oAuthToken).Get(gitHubUserURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not get user data")
	}
	defer oAuthUserInfoReq.Body.Close()
	var gUser struct {
		ID        int    `json:"id"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err = json.NewDecoder(oAuthUserInfoReq.Body).Decode(&gUser); err != nil {
		return nil, errors.Wrap(err, "decoding user info failed")
	}
	return &user{
		ID:      string(gUser.ID),
		Name:    gUser.Name,
		Picture: gUser.AvatarURL + "&s=64",
	}, nil
}

func (a *githubAdapter) GetOAuthProviderName() string {
	return "github"
}
