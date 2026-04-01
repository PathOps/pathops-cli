package auth

import "errors"

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
}

func LoginWithBrowserPKCE(issuer, clientID string) (LoginResult, error) {
	return LoginResult{}, errors.New("OIDC browser login not implemented yet")
}