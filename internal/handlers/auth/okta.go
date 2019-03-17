package auth

import (
        "context"
        "encoding/json"
        "net/url"
        "strings"

        "github.com/mxschmitt/golang-url-shortener/internal/util"
        "github.com/sirupsen/logrus"

        "github.com/pkg/errors"
        "golang.org/x/oauth2"
)

type oktaAdapter struct {
        config *oauth2.Config
}

// NewOktaAdapter creates an oAuth adapter out of the credentials and the baseURL
func NewOktaAdapter(clientID, clientSecret, endpointURL string) Adapter {

        if endpointURL == "" {
                logrus.Error("Configure Okta Endpoint")
        }

        return &oktaAdapter{&oauth2.Config{
                ClientID:     clientID,
                ClientSecret: clientSecret,
                RedirectURL:  util.GetConfig().BaseURL + "/api/v1/auth/okta/callback",
                Scopes: []string{
                        "profile",
                        "openid",
                        "offline_access",
                },
                Endpoint: oauth2.Endpoint{
                        AuthURL:  endpointURL + "/v1/authorize",
                        TokenURL: endpointURL + "/v1/token",
                },
        }}
}

func (a *oktaAdapter) GetRedirectURL(state string) string {
        return a.config.AuthCodeURL(state)
}

func (a *oktaAdapter) GetUserData(state, code string) (*user, error) {

        logrus.Debugf("Getting User Data with state: %s, and code: %s", state, code)
        oAuthToken, err := a.config.Exchange(context.Background(), code)
        if err != nil {
                return nil, errors.Wrap(err, "could not exchange code")
        }
        if util.GetConfig().Okta.EndpointURL == "" {
                logrus.Error("Okta EndpointURL is Empty")
        }
        oktaUrl, err := url.Parse(util.GetConfig().Okta.EndpointURL)
        if err != nil {
                return nil, errors.Wrap(err, "could not parse Okta EndpointURL")
        }
        oktaBaseURL := strings.Replace(oktaUrl.String(), oktaUrl.RequestURI(), "", 1)
        oAuthUserInfoReq, err := a.config.Client(context.Background(), oAuthToken).Get(oktaBaseURL + "/api/v1/users/me")
        if err != nil {
                return nil, errors.Wrap(err, "could not get user data")
        }
        defer oAuthUserInfoReq.Body.Close()
        var oUser struct {
                ID   int    `json:"sub"`
                // Custom URL property for user Avatar can go here
                Name string `json:"name"`
        }
        if err = json.NewDecoder(oAuthUserInfoReq.Body).Decode(&oUser); err != nil {
                return nil, errors.Wrap(err, "decoding user info failed")
        }
        return &user{
                ID:      string(oUser.ID),
                Name:    oUser.Name,
                Picture: util.GetConfig().BaseURL + "/images/okta_logo.png", // Default Okta Avatar
        }, nil
}

func (a *oktaAdapter) GetOAuthProviderName() string {
        return "okta"
}
