package auth

import (
	"context"
	"strings"

	"github.com/mxschmitt/golang-url-shortener/internal/util"
	"github.com/sirupsen/logrus"

	oidc "github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type genericOIDCAdapter struct {
	config   *oauth2.Config
	oidc     *oidc.Config
	provider *oidc.Provider
}

type claims struct {
	PreferredUsername string `json:"sub"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	ACR               string `json:"acr"`
}

// NewGenericOIDCAdapter creates an oAuth adapter out of the credentials and the baseURL
func NewGenericOIDCAdapter(clientID, clientSecret, endpointURL string) Adapter {
	endpointURL = strings.TrimSuffix(endpointURL, "/")

	if endpointURL == "" {
		logrus.Error("Configure GenericOIDC Endpoint")
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, endpointURL)
	if err != nil {
		logrus.Error("Configure GenericOIDC Endpoint: " + err.Error())
	}

	redirectURL := util.GetConfig().BaseURL + "/api/v1/auth/generic_oidc/callback"
	// Configure an OpenID Connect aware OAuth client.
	return &genericOIDCAdapter{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			// Discovery returns the OAuth endpoints.
			Endpoint: provider.Endpoint(),
			// "openid" is a required scope for OpenID Connect flows.
			Scopes: []string{
				"profile",
				"openid",
				"offline_access",
			},
		},
		oidc: &oidc.Config{
			ClientID: clientID,
		},
		provider: provider,
	}
}

func (a *genericOIDCAdapter) GetRedirectURL(state string) string {
	return a.config.AuthCodeURL(state)
}

func (a *genericOIDCAdapter) GetUserData(state, code string) (*user, error) {

	logrus.Debugf("Getting User Data with state: %s, and code: %s", state, code)
	oAuthToken, err := a.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, errors.Wrap(err, "could not exchange code")
	}

	rawIDToken, ok := oAuthToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.Wrap(err, "No id_token field in oauth2 token.")
	}

	idToken, err := a.provider.Verifier(a.oidc).Verify(context.Background(), rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "Something went wrong verifying the token: "+err.Error())
	}

	var oUser claims
	if err = idToken.Claims(&oUser); err != nil {
		return nil, errors.Wrap(err, "Something went wrong verifying the token: "+err.Error())
	}

	return &user{
		ID:      string(oUser.PreferredUsername),
		Name:    oUser.Name,
		Picture: util.GetConfig().BaseURL + "/images/generic_oidc_logo.png", // Default GenericOIDC Avatar
	}, nil
}

func (a *genericOIDCAdapter) GetOAuthProviderName() string {
	return "generic_oidc"
}
