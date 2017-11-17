package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/microsoft"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type microsoftAdapter struct {
	config *oauth2.Config
}

// NewMicrosoftAdapter creates an oAuth adapter out of the credentials and the baseURL
func NewMicrosoftAdapter(clientID, clientSecret, baseURL string) Adapter {
	return &microsoftAdapter{&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  baseURL + "/api/v1/auth/microsoft/callback",
		Scopes: []string{
			"wl.basic",
		},
		Endpoint: microsoft.LiveConnectEndpoint,
	}}
}

func (a *microsoftAdapter) GetRedirectURL(state string) string {
	return a.config.AuthCodeURL(state)
}

func (a *microsoftAdapter) GetUserData(state, code string) (*user, error) {
	logrus.Debugf("Getting User Data with state: %s, and code: %s", state, code)
	oAuthToken, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.Wrap(err, "could not exchange code")
	}
	oAuthUserInfoReq, err := a.config.Client(context.Background(), oAuthToken).Get("https://apis.live.net/v5.0/me")
	if err != nil {
		return nil, errors.Wrap(err, "could not get user data")
	}
	defer oAuthUserInfoReq.Body.Close()
	var mUser struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err = json.NewDecoder(oAuthUserInfoReq.Body).Decode(&mUser); err != nil {
		return nil, errors.Wrap(err, "decoding user info failed")
	}
	return &user{
		ID:      mUser.ID,
		Name:    mUser.Name,
		Picture: fmt.Sprintf("https://apis.live.net/v5.0/%s/picture", mUser.ID),
	}, nil
}

func (a *microsoftAdapter) GetOAuthProviderName() string {
	return "microsoft"
}
