package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Tokens struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	IDToken      string    `json:"idToken"`
	TokenType    string    `json:"tokenType"`
	Expiry       time.Time `json:"expiry"`
	Issuer       string    `json:"issuer"`
	ClientID     string    `json:"clientId"`
}

func tokenPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "pathops", "tokens.json"), nil
}

func SaveTokens(tokens Tokens) error {
	path, err := tokenPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func LoadTokens() (Tokens, error) {
	path, err := tokenPath()
	if err != nil {
		return Tokens{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Tokens{}, err
	}

	var tokens Tokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return Tokens{}, err
	}

	return tokens, nil
}

func HasTokens() bool {
	_, err := LoadTokens()
	return err == nil
}

func LoadAccessToken() (string, error) {
	t, err := LoadTokens()
	if err != nil {
		return "", err
	}
	if t.AccessToken == "" {
		return "", errors.New("missing access token")
	}
	return t.AccessToken, nil
}

func IsAccessTokenExpired(t Tokens, skew time.Duration) bool {
	if t.Expiry.IsZero() {
		return false
	}
	return time.Now().Add(skew).After(t.Expiry)
}
