package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	Expiry       time.Time
}

type oidcDiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

type callbackResult struct {
	Code  string
	State string
	Err   string
}

func LoginWithBrowserPKCE(issuer, clientID string) (LoginResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	discovery, err := discoverOIDC(ctx, issuer)
	if err != nil {
		return LoginResult{}, err
	}

	verifier, challenge, err := NewPKCE()
	if err != nil {
		return LoginResult{}, err
	}

	state, err := randomURLSafeString(24)
	if err != nil {
		return LoginResult{}, err
	}

	listener, redirectURI, err := startLoopbackListener()
	if err != nil {
		return LoginResult{}, err
	}
	defer listener.Close()

	callbackCh := make(chan callbackResult, 1)

	server := &http.Server{
		Handler: callbackHandler(callbackCh),
	}
	defer server.Close()

	go func() {
		_ = server.Serve(listener)
	}()

	conf := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURI,
		Scopes:      []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
	}

	authURL := conf.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	if err := OpenBrowser(authURL); err != nil {
		return LoginResult{}, fmt.Errorf("could not open browser automatically: %w", err)
	}

	fmt.Printf("Opening browser for login...\n")
	fmt.Printf("If the browser does not open, visit:\n%s\n\n", authURL)

	var cb callbackResult

	select {
	case <-ctx.Done():
		return LoginResult{}, errors.New("login timed out waiting for browser callback")
	case cb = <-callbackCh:
	}

	_ = server.Shutdown(context.Background())

	if cb.Err != "" {
		return LoginResult{}, fmt.Errorf("authorization failed: %s", cb.Err)
	}

	if cb.State != state {
		return LoginResult{}, errors.New("invalid oauth state")
	}

	token, err := conf.Exchange(
		ctx,
		cb.Code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return LoginResult{}, fmt.Errorf("token exchange failed: %w", err)
	}

	idToken, _ := token.Extra("id_token").(string)

	return LoginResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      idToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

func discoverOIDC(ctx context.Context, issuer string) (oidcDiscoveryDocument, error) {
	wellKnown := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return oidcDiscoveryDocument{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return oidcDiscoveryDocument{}, fmt.Errorf("oidc discovery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return oidcDiscoveryDocument{}, fmt.Errorf("oidc discovery failed: status=%d", resp.StatusCode)
	}

	var doc oidcDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return oidcDiscoveryDocument{}, fmt.Errorf("invalid oidc discovery document: %w", err)
	}

	if doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" {
		return oidcDiscoveryDocument{}, errors.New("oidc discovery document missing endpoints")
	}

	return doc, nil
}

func startLoopbackListener() (net.Listener, string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("could not open local callback listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d", port)

	return listener, redirectURI, nil
}

func callbackHandler(ch chan<- callbackResult) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		res := callbackResult{
			Code:  q.Get("code"),
			State: q.Get("state"),
			Err:   q.Get("error"),
		}

		select {
		case ch <- res:
		default:
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, successHTML())
	})

	return mux
}

func successHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>PathOps Login</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body {
      font-family: sans-serif;
      max-width: 640px;
      margin: 60px auto;
      padding: 0 16px;
      line-height: 1.5;
    }
    .card {
      border: 1px solid #ddd;
      border-radius: 12px;
      padding: 24px;
    }
  </style>
</head>
<body>
  <div class="card">
    <h1>PathOps login completed</h1>
    <p>You can close this tab and return to the terminal.</p>
  </div>
</body>
</html>`
}

func randomURLSafeString(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func RefreshToken(ctx context.Context, issuer, clientID, refreshToken string) (LoginResult, error) {
	discovery, err := discoverOIDC(ctx, issuer)
	if err != nil {
		return LoginResult{}, err
	}

	conf := &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
	}

	src := conf.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	token, err := src.Token()
	if err != nil {
		return LoginResult{}, fmt.Errorf("refresh token exchange failed: %w", err)
	}

	idToken, _ := token.Extra("id_token").(string)

	return LoginResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      idToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}, nil
}

func BuildManualBrowserURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return u.String()
}
