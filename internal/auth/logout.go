package auth

import (
	"fmt"
	"net/url"
	"strings"
)

func BuildLogoutURL(issuer string) (string, error) {
	if strings.TrimSpace(issuer) == "" {
		return "", fmt.Errorf("issuer is required")
	}

	base := strings.TrimRight(issuer, "/")
	logoutURL := base + "/protocol/openid-connect/logout"

	u, err := url.Parse(logoutURL)
	if err != nil {
		return "", fmt.Errorf("invalid logout url: %w", err)
	}

	return u.String(), nil
}
